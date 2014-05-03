package control

import (
	"errors"
	"fmt"
	"github.com/mrmorphic/goss/data"
	"github.com/mrmorphic/goss/orm"
	"github.com/mrmorphic/goss/template"
	"net/http"
	"strconv"
	"strings"
)

type ContentController interface {
	// used to set the DataObject for a content controller
	SetObject(interface{})

	// set the context object for rendering. This is required because there is no inheritance; calling
	// ServerHTTP on a controller that embeds ContentControllerStruct gets the embedded object, not
	// the output object, so we can't pick up symbol references from upstream.
	SetContext(interface{})
}

// ContentController is intended as a simple page-based controller. When a DataObject is mapped to this controller,
// it will render by locating the templates for the page type and rendering using the templating engine. Typically
// ContentController is embedded into other types that implement functions specific to that type that can be used
// in the templates.
type ContentControllerStruct struct {
	BaseController

	context interface{}

	Fallback interface{}
}

func (c *ContentControllerStruct) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	className := data.Eval(c.GetObject(), "ClassName").(string)
	templates := []string{"Page", className}
	e := template.RenderWith(w, templates, c.context, nil)

	if e != nil {
		ErrorHandler(w, e)
		return
	}
}

func (c *ContentControllerStruct) SetContext(ctx interface{}) {
	c.context = ctx
}
func (c *ContentControllerStruct) SetObject(obj interface{}) {
	c.Fallback = obj
}

func (c *ContentControllerStruct) GetObject() interface{} {
	return c.Fallback
}

// Given a request, follow the segments through sitetree to find the page that is being requested. Doesn't
// understand actions, so just finds the page. Returns ID of SiteTree_Live record or 0 if it can't find a
// matching page.
// @todo Understand BaseController actions, or break on the furthest it gets up the tree
// @todo cache site tree
func findPageToRender(r *http.Request) (int, error) {
	s := strings.Trim(r.URL.Path, "/")
	path := strings.Split(s, "/")

	if len(path) == 0 || path[0] == "" {
		// find a home page ID
		r, e := orm.Query("select \"ID\" from \"SiteTree_Live\" where \"URLSegment\"='home' and \"ParentID\"=0")
		defer r.Close()

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
		r, e := orm.Query("select \"ID\",\"ParentID\" from \"SiteTree_Live\" where \"URLSegment\"='" + p + "' and \"ParentID\"=" + strconv.Itoa(currParentID))
		defer r.Close()

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
func findNotFoundPage(r *http.Request) (int, error) {
	return 0, errors.New("not found")
}

// Handle a request for a general page out of site tree:
// - pull apart the path, and use it to guide the location of a site tree record
//   from the SS DB, matching URL segments exactly
// - if there is no matching page, find an error page instead
// - with the page in sitetree located, use ClassName to determine the controller that should be invoked.
// - grab the data object and render the template with it.
func SiteTreeHandler(w http.ResponseWriter, r *http.Request) {
	pageID, e := findPageToRender(r)
	if e != nil {
		ErrorHandler(w, e)
		return
	}

	if pageID == 0 {
		pageID, e = findNotFoundPage(r)
	}

	if e != nil {
		ErrorHandler(w, e)
		return
	}

	if pageID == 0 {
		// uh oh, couldn't find anything we could render off in site tree
		e = errors.New("Could not find anything to render at all")
		ErrorHandler(w, e)
		return
	}

	//	fmt.Printf("SiteTreeHandler has found a page: %d\n", pageID)

	q := orm.NewQuery("SiteTree").Where("\"SiteTree_Live\".\"ID\"=" + strconv.Itoa(pageID))
	v, _ := q.Run()

	if e != nil {
		ErrorHandler(w, e)
		return
	}

	res := v.(orm.DataList)
	items, e := res.Items()
	if e != nil {
		ErrorHandler(w, e)
	}

	if len(items) == 0 {
		e = errors.New("Could not locate object with ID " + strconv.Itoa(pageID))
		ErrorHandler(w, e)
		return
	}

	page := items[0]

	renderWithMatchedController(w, r, page)
}

// Given a page, find a controller that says it can handle it, and render the page with that.
func renderWithMatchedController(w http.ResponseWriter, r *http.Request, page interface{}) {
	// locate a controller
	className := data.Eval(page, "ClassName").(string)
	c, e := getControllerInstance(className)

	if e != nil {
		ErrorHandler(w, e)
		return
	}

	c.Init(r)

	// if the controller is a ContentController then set the object.
	if cc, ok := c.(ContentController); ok {
		cc.SetObject(page)
		cc.SetContext(c)
	}

	//	fmt.Printf("after init c is %s\n", c)
	c.ServeHTTP(w, r)
}

// If we get an error that can't be handled, call this to write the response
func ErrorHandler(w http.ResponseWriter, e error) {
	http.Error(w, fmt.Sprintf("%s", e), http.StatusBadRequest)
}

func AssetHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Trying to get assets")
}
