package goss

import (
	"net/http"
	"database/sql"
	"fmt"
	"errors"
	"strings"
	"strconv"
)

type DBConnFactory func()(*sql.DB, error)
type DBCloseConn func(*sql.DB)
type RenderCallback func(http.ResponseWriter, *http.Request, *RequestContext, *DataObject)

type NavigationProvider interface {
	Menu(level int) (DataList)
}

var dbFactory DBConnFactory
var dbClose DBCloseConn
var renderFunction RenderCallback

type RequestContext struct {
	db *sql.DB
	metadata *DBMetadata
}

// Given a request, follow the segments through sitetree to find the page that is being requested. Doesn't
// understand actions, so just finds the page. Returns ID of SiteTree_Live record or 0 if it can't find a
// matching page.
// @todo Understand controller actions, or break on the furthest it gets up the tree
// @todo cache site tree
func (ctx *RequestContext) findPageToRender(r *http.Request) (int, error) {
	s := strings.Trim(r.URL.Path, "/")
	path := strings.Split(s, "/")

	currParentID := 0
	for _,p := range path {
		r, e := ctx.Query("select ID,ParentID from SiteTree_Live where URLSegment='" + p + "' and ParentID=" + strconv.Itoa(currParentID))
		if e != nil {
			return 0, e
		}
		if !r.Next() {
			return 0, nil
		}

		var ID, ParentID int
		e = r.Scan(&ID, &ParentID)
		currParentID = ID
	}

	// if we get to the end, we've found a matching ID in SiteTree_Live
	return currParentID, nil
}

// Find the 'PageNotFound' error page and return it's ID
func (ctx *RequestContext) findNotFoundPage(r *http.Request) (int, error) {
	return 0, errors.New("not found")
}

func (ctx *RequestContext) Menu(level int) (set *DataList, e error) {
	q := NewQuery().BaseClass("SiteTree").Where("ParentID=0").Where("ShowInMenus=1").OrderBy("\"Sort\" ASC")
	set, e = q.Exec(ctx)
	return
}

func (ctx *RequestContext) SiteConfig() (obj *DataObject, e error) {
	q := NewQuery().BaseClass("SiteConfig").Limit(0,1)
	res, e := q.Exec(ctx)
	if e != nil {
		return nil, e
	}

	if len(res.Items) < 1 {
		return nil, errors.New("There is no SiteConfig record")
	}

	return res.Items[0], nil
}

/**
 * Given a data object, return a path to it. For a site object, this is the URLSegments of the
 * descendents and the object. For other (non-hierarchical objects) this is an empty string, because
 * there is no structure. Note this is not a complete Link, because the link is a function
 * of the presentation layer, not the model. But this is helpful especially for SiteTree objects.
 */
func (ctx *RequestContext) Path(obj *DataObject, field string) (string, error) {
	hier, e := ctx.metadata.IsHierarchical(obj.AsString("ClassName"))
	if e != nil {
		return "", e
	}
	if !hier {
		return "", nil
	}

//	i,_ := obj.AsInt("ParentID")
//	fmt.Printf("ParentID for object is %d\n", i)
//fmt.Printf("Object is %s\n", obj)
	res := obj.AsString(field)
	for parentID, e := obj.AsInt("ParentID"); e != nil && parentID > 0; {
		// @todo not BaseClass("SiteTree"), derive the base class using metadata.
		q := NewQuery().BaseClass("SiteTree").Where("ID=" + strconv.Itoa(parentID))
		ds, e := q.Exec(ctx)
		if e != nil {
			return "", e
		}
		obj = ds.First()
		res = obj.AsString(field) + "/" + res
	}
	return res, nil
}

/**
 * Set the DB factory and DB close connection methods. goss does not know how to connect to the DB.
 * This may change, as do need in the orm to know what kind of DB wr're dealing with for SQL generation.
 */
func SetConnection(factory DBConnFactory, closeConn DBCloseConn) {
	dbFactory = factory
	dbClose = closeConn
}

/**
 * Set the render callback method. This is used when we've identified a site tree record to render.
 */
func SetRenderCallback(renderFn RenderCallback) {
	renderFunction = renderFn
}

/**
 * Set the metadata that was used to generate the database we're connecting with. We need this
 * because we don't have direct access to this from the PHP app, and we need it to understand
 * class hierarchy, class properties and their types.
 */
func SetMetadata() {

}

// Handle a request for a general page out of site tree:
// - pull apart the path, and use it to guide the location of a site tree record
//   from the SS DB, matching URL segments exactly
// - if there is no matching page, find an error page instead
// - with the page in sitetree located, use ClassName to find a template
// - grab the data object and render the template with it.
//
// @todo find site tree for matches
// @todo find error page if no match
// @todo locate template from classname
// @todo read sitetree for identified record. Needs to read all properties of the site tree. How?
// @todo read full object from site tree, which requires meta-data from SS application.
func SiteTreeHandler(w http.ResponseWriter, r *http.Request) {
	db, e := dbFactory()
	if e != nil {
		ErrorHandler(w, r, e)
		return
	}
	defer dbClose(db)

	metadata := new(DBMetadata)
	ctx := &RequestContext{db, metadata}

	pageID, e := ctx.findPageToRender(r)
	if e != nil {
		ErrorHandler(w, r, e)
		return
	}

	if pageID == 0 {
		pageID, e = ctx.findNotFoundPage(r)
	}

	if e != nil {
		ErrorHandler(w, r, e)
		return
	}

	if pageID == 0 {
		// uh oh, couldn't find anything we could render off in site tree
		e = errors.New("Could not find anything to render at all")
		ErrorHandler(w, r, e)
		return
	}

//	fmt.Printf("SiteTreeHandler has found a page: %d\n", pageID)

	q := NewQuery().BaseClass("SiteTree").Where("\"ID\"=" + strconv.Itoa(pageID))
	res, _ := q.Exec(ctx)

	if e != nil {
		ErrorHandler(w, r, e)
		return		
	}

	if len(res.Items) == 0 {
		e  = errors.New("Could not locate object with ID " + strconv.Itoa(pageID))
		ErrorHandler(w, r, e)
		return		
	}

	page := res.Items[0]

	renderFunction(w, r, ctx, page)
}

// If we get an error that can't be handled, call this to write the response
func ErrorHandler(w http.ResponseWriter, r *http.Request, e error) {
	fmt.Fprintf(w, "Error loading page: %s", e)
}

func AssetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Trying to get assets")
}
