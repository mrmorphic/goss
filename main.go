package goss

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type RenderFunc func(http.ResponseWriter, *http.Request, *DBContext, *DataObject)

type NavigationProvider interface {
	Menu(level int) DataList
}

// database is actually a connection pool. The pool is automatically managed, and works across go-routines.
var database *sql.DB

// metadataSource is a path to the file containing metadata used by the ORM.
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

type setupFunc func(conf Config) error

// SetConfig tells goss the configuration object to use. This should be called before requests are accepted.
// The configuration properties that goss understands will be read at this point.
func SetConfig(conf Config) error {
	setupFunctions := []func(Config) error{setupDB, setupMetadata}
	for _, fn := range setupFunctions {
		e := fn(conf)
		if e != nil {
			return e
		}
	}

	return nil
}

// setupDB creates the database connection pool. This is shared across go-routines for all requests,
// and the pool management is managed automatically by the sql package.
func setupDB(config Config) error {
	// Get the properties we expect.
	driverName := config.AsString("goss.database.driverName")
	if driverName == "" {
		return errors.New("goss requires config property goss.database.driverName to be set.")
	}

	dataSourceName := config.AsString("goss.database.dataSourceName")
	if dataSourceName == "" {
		return errors.New("goss requires config property goss.database.dataSourceName to be set.")
	}

	maxIdleConnections := -1 // default is no idle connections
	mi := config.Get("goss.database.maxIdleConnections")
	maxIdleConnections, ok := mi.(int)
	if !ok {
		return errors.New("goss expects config property goss.database.maxIdleConnections to be of type 'int'.")
	}

	// put back in once at go 1.2
	// maxOpenConnections := -1 // default is no limit on open connections
	// mo := config.Get("goss.database.maxOpenConnections")
	// maxOpenConnections, ok = mo.(int)
	// if !ok {
	// 	return errors.New("goss expects config property goss.database.maxOpenConnections to be of type 'int'.")
	// }

	database, e := sql.Open(driverName, dataSourceName)
	if e != nil {
		return e
	}

	fmt.Printf("opened database %s: %s\n", driverName, dataSourceName)

	database.SetMaxIdleConns(maxIdleConnections)
	//	database.SetMaxOpenConns(maxOpenConnections) // requires go 1.2

	return nil
}

func setupMetadata(conf Config) error {
	// metadataSource = conf.AsString("goss.metadata")
	// if metadataSource == "" {
	// 	return errors.New("goss requires configuration property goss.metadata is set.")
	// }
	return nil
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
	if dbMetadata == nil {
		// Generate the metadata object on demand
		dbMetadata = new(DBMetadata)
	}
	e := dbMetadata.RefreshOnDemand(metadataSource)
	if e != nil {
		ErrorHandler(w, r, e)
		return
	}

	// @todo can metadata be a shared global instead of allocating it each time? What about updates to the metadata?
	// @todo use metadataSource to initialise DBMetadata, if the file hasn't changed.
	ctx := &DBContext{database, dbMetadata}

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
