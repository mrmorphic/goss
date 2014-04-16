package goss

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type DBConnFactory func() (*sql.DB, error)
type DBCloseConn func(*sql.DB)
type RenderFunc func(http.ResponseWriter, *http.Request, *DBContext, *DataObject)

type NavigationProvider interface {
	Menu(level int) DataList
}

var dbFactory DBConnFactory
var dbClose DBCloseConn
var metadataSource string

var dbMetadata *DBMetadata

type DBContext struct {
	db       *sql.DB
	Metadata *DBMetadata
}

// Given a request, follow the segments through sitetree to find the page that is being requested. Doesn't
// understand actions, so just finds the page. Returns ID of SiteTree_Live record or 0 if it can't find a
// matching page.
// @todo Understand BaseController actions, or break on the furthest it gets up the tree
// @todo cache site tree
func (ctx *DBContext) findPageToRender(r *http.Request) (int, error) {
	s := strings.Trim(r.URL.Path, "/")
	path := strings.Split(s, "/")

	if len(path) == 0 || path[0] == "" {
		// find a home page ID
		r, e := ctx.Query("select \"ID\" from \"SiteTree_Live\" where \"URLSegment\"='home' and \"ParentID\"=0")
		if e != nil {
			return 0, e
		}
		if !r.Next() {
			return 0, nil
		}

		var ID int
		e = r.Scan(&ID)

		return ID, e
	}

	currParentID := 0
	for _, p := range path {
		r, e := ctx.Query("select \"ID\",\"ParentID\" from \"SiteTree_Live\" where \"URLSegment\"='" + p + "' and \"ParentID\"=" + strconv.Itoa(currParentID))
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
func (ctx *DBContext) findNotFoundPage(r *http.Request) (int, error) {
	return 0, errors.New("not found")
}

func SetConfig(conf Config) {
}

// SetConnection is used by the application to provide the DB factory and DB close connection methods.
// goss does not know how to connect to the DB by itself.
// This may change, as do need in the orm to know what kind of DB we're dealing with for SQL generation.
func SetConnection(factory DBConnFactory, closeConn DBCloseConn, metadataFile string) {
	dbFactory = factory
	dbClose = closeConn
	metadataSource = metadataFile
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
func SiteTreeHandler(w http.ResponseWriter, r *http.Request, renderFn RenderFunc) {
	db, e := dbFactory()
	if e != nil {
		ErrorHandler(w, r, e)
		return
	}
	defer dbClose(db)

	if dbMetadata == nil {
		// Generate the metadata object on demand
		dbMetadata = new(DBMetadata)
	}
	e = dbMetadata.RefreshOnDemand(metadataSource)
	if e != nil {
		ErrorHandler(w, r, e)
		return
	}

	// @todo can metadata be a shared global instead of allocating it each time? What about updates to the metadata?
	// @todo use metadataSource to initialise DBMetadata, if the file hasn't changed.
	ctx := &DBContext{db, dbMetadata}

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

	q := NewQuery("SiteTree").Where("\"SiteTree_Live\".\"ID\"=" + strconv.Itoa(pageID))
	res, _ := q.Exec(ctx)

	if e != nil {
		ErrorHandler(w, r, e)
		return
	}

	if len(res.Items) == 0 {
		e = errors.New("Could not locate object with ID " + strconv.Itoa(pageID))
		ErrorHandler(w, r, e)
		return
	}

	page := res.Items[0]

	renderFn(w, r, ctx, page)
}

// If we get an error that can't be handled, call this to write the response
func ErrorHandler(w http.ResponseWriter, r *http.Request, e error) {
	fmt.Fprintf(w, "Error loading page: %s", e)
}

func AssetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Trying to get assets")
}
