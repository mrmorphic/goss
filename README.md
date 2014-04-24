# goss

## Overview

Goss is an experimental library to interface Go applications to SilverStripe applications. Key features are:

 *	An ORM that reads from SilverStripe databases directly (SS 3+)
 *	Base controller that provides common functions for construction of simple web sites.
 *	Templating implementation to re-use SilverStripe templates

Typical use cases envisioned for goss are:

 *	Construction of web service APIs against the SilverStripe database that
	need to be very fast.
 *	Construction of limited web site front ends that don't require all
    the features of the SilverStripe framework, but which will benefit from speed (e.g. constantly changing dynamic data)

The motivation for the library is to provide a way to build components within a SilverStripe installation where high performance is critical. Rough tests indicate approximately 2 orders of magnitude speed improvement for some
simple functions implemented in go vs in PHP/SilverStripe on apache.

goss has a number of sub-packages to implement various components. While these components are designed to work well together, they are also decoupled so you can provide alternative implementations. For example:

 *	You can use the Go runtimes templating system.
 *	You can provide alternative configuration systems.
 *	You can provide a different ORM implementation.
 *	You can bypass controllers and use handlers directly, including handlers
 	from other Go web frameworks.

## Author

Mark Stephens (mr.morphic@gmail.com)


## Status

This library is still is fairly early stages of development, and should be considered unstable. The following are known to work to some degree:

 *	ORM queries can be formulated and executed.
 *	Metadata from the database can be read and successfully used to generate queries.
 *	Query results can be used in templates.
 *	Basic controller operations are supported.
 *	SilverStripe templates can (mostly) be rendered.

Most features of SilverStripe are not implemented in goss. A few likely candidates for development include:

 *	Ability to issue queries in either live or staging (only live is supported at present)
 *	Support for object writes via ORM
 *	Limited session support
 *	Limited authentication support
 *	Support for access to functions within the SilverStripe application rather than accessing database directly (e.g.
 	re-use of business rule logic)

## Example Usage

NOTE: this section needs to be re-written. The entire mechanism described here has been replaced.

This is a simple annotated website using goss:

	import (
		"fmt"
		"net/http"
		"database/sql"
		_ "code.google.com/p/go-mysql-driver/mysql"
		"goss"
		"html/template"
	)

	func main() {
		// Tell goss where the metadata json file is located. See the ORM section for more information on where it
		// comes from.
		goss.SetConnection(openConn, closeConn, "/sites/mysite/assets/goss/metadata.json")

		// Set a rendering callback, using in conjunction with SiteTreeHandler, below
		goss.SetRenderCallback(RenderCallback)

		// Set up handlers for static stuff
		http.Handle("/css/", http.StripPrefix("/css", http.FileServer(http.Dir("./css"))))
		http.Handle("/js/", http.StripPrefix("/js", http.FileServer(http.Dir("./js"))))
		http.Handle("/images/", http.StripPrefix("/images", http.FileServer(http.Dir("./images"))))
		http.Handle("/thirdparty/", http.StripPrefix("/thirdparty", http.FileServer(http.Dir("./thirdparty"))))

		// All other URLs are handled by SiteTreeHandler, which will parse the URL against SiteTree in
		// a similar way to SilverStripe framework, and once it identifies the SiteTree object, invokes
		// RenderCallback, defined above.
		http.HandleFunc("/", goss.SiteTreeHandler)

		// Start listening
		http.ListenAndServe(":8081", nil)
	}

	var cacheddb *sql.DB

	// A function provided to goss to open connections
	func openConn() (db *sql.DB, e error) {
		if cacheddb != nil {
			return cacheddb, nil
		}

		db, e = sql.Open("mysql", "myuser:mypassword@tcp(127.0.0.1:3306)/my_ss_database")
		if e != nil {
			return nil, e
		}

		// goss uses ANSI compliant queries, so we need to set this, for mysql.
		_, e = db.Query("SET GLOBAL TRANSACTION ISOLATION LEVEL SERIALIZABLE;")
		_, e = db.Query("SET GLOBAL sql_mode = 'ANSI'")

		cacheddb = db
		return db, e
	}

	// A function provided to goss to close connections.
	func closeConn(db *sql.DB) {
		//	fmt.Println("closeConn")
		//	if db != nil {
		//		db.Close()
		//	}
	}

	var baseURL = "http://127.0.0.1:8081/"

	// Define a controller class for our own functions
	type GenericController struct {
		goss.BaseController
	}

	// Generate a link to a given DataObject. Note that this is very different from how it's implemented in
	// a SilverStripe site, where each DataObject implements Link. In Go, we don't have inheritance this way.
	func (c *GenericController) DataLink(obj *goss.DataObject) string {
		rel, _ := c.Path(obj, "URLSegment")
		return baseURL + rel
	}

	func (c *GenericController) HomeLink() string {
		return baseURL
	}

	func (c *GenericController) LinkByType(class string) string {
		ds, e := goss.NewQuery("SiteTree").Where("\"ClassName\"='" + class + "'").Exec(c.DB)
		if e != nil {
			fmt.Printf("error: %s\n", e)
		}
		obj := ds.First();
		return c.DataLink(obj)
	}

	func (c *GenericController) BaseURL() string {
		return baseURL
	}

	func RenderCallback(w http.ResponseWriter, r *http.Request, ctx *goss.DBContext, object *goss.DataObject) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")

		var ro goss.Controller

		// at this point, we want to create a custom render context for the type of page, or the default
		// so we have a switch on type, create an object and initialise it. Here we just look at the class name
		// of the object provided, and if it's my custom page type I create a controller of that type, otherwise
		// just a generic controller.

		switch object.AsString("ClassName") {
		case "MyPage":
			ro = new(MyPageController)
		default:
			ro = new(GenericController)
		}

		// Give the controller the context
		ro.Init(w, r, ctx, object)

		// Invoke go's templating engine to render an output.
		t, _ := template.ParseFiles("templates/" + object.AsString("ClassName") + ".html", "templates/base.html")
		t.ExecuteTemplate(w, "base", ro)
	}

Here is an example of base.html referred to above, with HTML simplified for brevity:

	{{define "base"}}<!DOCTYPE html>

	<html lang="en">
		<head>
	<base href="{{.BaseURL}}">
		<title>{{.SiteConfig.AsString "Title"}}</title>
		<meta http-equiv="Content-Type" content="text/html; charset=utf-8" />

		{{template "head" .}}
	</head>

	<body class="{{.Object.AsString "ClassName"}}">
		<div class="container">
			<div class="row">
				<div class="span10">
					<a href="{{.BaseURL}}" class="brand" rel="home">
						<h1>{{.SiteConfig.AsString "Title"}}</h1>
						<p>{{.SiteConfig.AsString "Tagline"}}</p>
					</a>
				</div>
				<div class="span2">
					<div class="account-panel">
					{{if .CurrentMember}}
						Hi {{.CurrentMember.AsString "FirstName"}}(<a href="$LogoutLink">log out</a>)
					{{else}}
						<a href="$LoginLink">log in</a>
					{{end}}
					</div>
				</div>
			</div>

			<div class="navbar">
				<div class="navbar-inner">
					<ul class="nav">
						{{$base := .}}
						{{with .Menu 1}}{{range .Items}}
							<li class="$LinkingMode"><a href="{{$base.DataLink .}}" title="$Title.XML">{{if .AsString "MenuTitle"}}{{.AsString "MenuTitle"}}{{else}}{{.AsString "Title"}}{{end}}</a></li>
						{{end}}{{end}}
					</ul>
				</div>
			</div>

			<div class="main row" role="main">
				<div class="inner typography span12">
					{{template "body" .}}
				</div>
			</div>

		</div>
	</body>

	</html>
	{{end}}

And this is what the page-type specific template looks like:

	{{define "head"}}<title>Some Title</title>{{end}}
	{{define "body"}}
		{{with .Object}}
		<h1>{{.AsString "Title"}}</h1>
		{{.AsString "Content"}}
		{{end}}
	{{end}}

## Configuration

Configuration is provided to goss using the ConfigProvider interface. The package goss/config package provides an implementation of this interface, which you can create and use to read configuration from a file, as follows:

	import (
		"github.com/mrmorphic/goss/config"
	)

	// read from config.json.
	conf, e := config.ReadFromFile("config.json")
	if e != nil {
		log.Fatal(e)
		return
	}
	// Give goss the configuration object.
	e = goss.SetConfig(conf)
	if e != nil {
		log.Fatal(e)
		return
	}

	// you can put your own settings in here too
	u, e = url.Parse(conf.AsString("app.myAppName"))

ConfigProvider assumes that the configuration data is organised in a name space. The goss package expects certain properties under 'goss' at the top level:

 *	goss.database.driverName: driver name as required by sql.Open
 *	goss.database.dataSourceName: data source as required by sql.Open
 *	goss.database.maxIdleConnections: maximum number of idle connections in
	database pool provided by sql package.
 *	goss.database.maxOpenConnections: maximum number of open connections in
	the database pool provided by sql package. You need to ensure that this
	number plus the maximum number of connections of the SilverStripe site do
	not exceed the maximum number of connections of the database server
	itself, or you may get errors under load.
 *	goss.ssroot: the path to webroot of your SilverStripe installation.
 *	goss.siteUrl: the URL the site is publicly known as
 *	goss.defaultProtocol: must be http or https, and can be used in generation
	of URLs.
 *	goss.theme: the name of the theme for template rendering. There is a
	limitation in the goss templating engine that all templates must be
	located in the same theme.
 *	goss.metadata: a path to the metadata JSON file that contains metadata
 	for the ORM. This is typically automatically generated using the
 	github.com/mrmorphic/silverstripe-goss module.

## ORM

## Controllers

NOTE: this section needs to be rewritten. Controllers have been completely refactored.

Goss provides a Controller type which you can use to build your own page-specific controllers. This works quite
differently from controllers in the SilverStripe framework, and provides some of the functions that are actually
present in ViewableData, since the DataObject type in Goss is only a simple value container.

Typical use of Controller is as follows:

	// Define a controller class for our own functions
	type GenericController struct {
		goss.BaseController
	}

Functions that are specific to this type of page can be added here. Note, however, that you cannot override
methods from the base controller. This is a Go language constraint.

Some functions provided by Controller include:

 *	func (c *BaseController) Init(w http.ResponseWriter, r *http.Request, ctx *DBContext, object *DataObject)
	This must be called after constructing a controller object, and before rendering with it. It sets the context.

 *	func (ctl *BaseController) Menu(level int) (set *DataList, e error)
 	This is a helper function that attempts to behave like ViewableData::Menu. It currently only returns top level
 	menu items.

 *	func (ctl *BaseController) SiteConfig() (obj *DataObject, e error)
 	This is a helper function that returns the SiteConfig object.

 *	func (ctl *BaseController) Path(obj *DataObject, field string) (string, error)
 	This is a helper function that returns a portion of the path to a data object in SiteTree, by concatenating
 	the URLSegments. This is useful in writing link-generation functions on SiteTree objects.

## Templates

The template package implements the SilverStripe templating language. The intention is that templates may be developed that are used by both the SilverStripe host app as well as the goss app. Minor alterations may need to be made for templates that are to work in both environments.

As much as possible, the syntax has been made identical to the SilverStripe templating language. The main differences are going to be in templating features, such as $CurrentMember or $Children, which are methods provided by the underlying controller, as goss controllers will not have all of these.

### Implemented

This section lists features of the SilverStripe templating language that have been implemented. Some of the implementations may vary because of underlying differences in the systems.

 *	variable substitutions: $foo, {$foo}
 *	function substitutions: $foo(args), {$foo(args)}, $foo(arg1, arg2)
 *	<% if cond %>...<% end_if%> , <% else %> variation
 *	operators: ||, &&, ==, !=, >, >=, <, <=, not
 *	numeric literals (no decimals)
 *	string literals (backslashes in strings not handled)
 *	chained variable and functions, eg $foo.bar, $foo().bar("abc")
 *	main templates and main/layout arrangements, $Layout
 *	comments
 *	<% base_tag %>
 
### Not implemented

The following are parsed only:
 *	<% loop expr %>...<% end_loop %>
 *	<% with expr %>...<% end_with %>
 *	<% include %>, but not optional binding syntax


The following are not implemented. They are listed in approximate priority order for implementaton.

 *	else_if
 *	<% require ... %>
 *	\$var
 *	requirements injection
 *	shortcode handling
 *	backslash handling in string literals
 *	<%t ... %>
 *	<% include %> allows for an optional binding syntax for the included
	template. This extra syntax is not implemented.
 *	deprecated syntax of using identifiers without $ or double quotes
 *	<% cached %> blocks are parsed and handled correctly semantically, but
	there is no caching of the fragments, and any expressions in the
	<% cached %> tag are not evaluated.

## Revised Notes

 *  Muxer is used to map routes to handlers
 *  SiteTreeHandler is a handler
 *  Since DataObjects are thin and don't have methods attached in this release, controllers are used to do this
    as well. So caller needs to add mapping of DataObjects to controllers.
 *  route -> handler -> controller
 *  ContentController will automatically render based on templates.
