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

This is an example site (extremely simple) using goss:

	package main

	import (
		"fmt"
		_ "github.com/go-sql-driver/mysql"  // import the SQL driver
		"github.com/mrmorphic/goss"
		"github.com/mrmorphic/goss/config"
		"github.com/mrmorphic/goss/control"
		"log"
		"net/http"
	)

	func main() {
		// Read configuration from file system
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

		// Add controllers. These allow SiteTreeHandler to automatically
		// create a controller instance and get it to handle the request.
		control.AddController("HomePage", &HomePageController{})

		// add a rule that home page is handled by SiteTreeHandler.
		// @todo add assets and themes rules as well
		goss.AddMuxRule("^/$", func(w http.ResponseWriter, r *http.Request) {
			control.SiteTreeHandler(w, r)
		})

		http.HandleFunc("/", goss.MuxServe)
		fmt.Printf("==== about to listen\n")
		e = http.ListenAndServe(":8080", nil)
		if e != nil {
			log.Fatal(e)
		}
	}

	// A custom controller. It embeds ContentControllerStruct to get
	// 'inherited' behaviours, and adds a custom function for use in
	// the template
	type HomePageController struct {
		control.ContentControllerStruct
	}

	func (c *HomePageController) Salutation(name string) string {
		return "Dear " + name
	}

The config.json file might look like:

	{
		"goss":{
			"ssroot": "/var/www/demosite",
			"theme": "simple",
			"metadata": "/var/www/demosite/assets/goss/metadata.json",
			"siteUrl": "http://demosite.com/",
			"database": {
				"driverName": "mysql",
				"dataSourceName": "user:password@tcp(127.0.0.1:3306)/ss_demo",
				"maxIdleConnections": 10,
				"maxOpenConnections": 100
			}
		}
	}


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
 *	goss.cache.menuTTL: time-to-live in seconds of menu(n) cache. 0 means
	not cached.
 *	goss.cache.siteConfigTTL: time-to-live in seconds of site config cache.
 	0 means not cached.
 *	goss.cache.siteTreeTTL: time-to-live in seconds of site tree cache,
 	which is cache of some SiteTree properties, used for URL segment lookups.
 	0 means not cached.

## Objects and Interfaces

Go's object model is very different from PHP and SilverStripe's object model. For example, subcasses and extensions are used extensively in SilverStripe. Go uses interfaces extensively and composition of independent components, with
subclassing not being a language feature.

goss keeps as much as possible to how idiomatic Go works, while providing the conceptual framework of SilverStripe where it makes sense. To this end, goss works mostly with interface{}, which can be a value of any type. Concrete types are provided by the library as well.

Key interfaces of goss include:

 *	Evaluater - implementers are able to return a value for a
	variable reference or function call reference. Key implementers are
	DataObject, DataList, DefaultLocator (utility) and Controller. This is
	used extensively by the templating engine to resolve symbols, but is
	typically used by project code as "getters" for properties.
 *	ConfigurationProvider - implementers can provide config values
 *	RequirementsProvider - interface to a requirements back end, with a
 	default implementation of goss.requirements.DefaultRequirements

The current mapping from SilverStripe concepts to Go/goss:

 *	Controller - the request handling aspect of Controller is a direct
 	parallel to net/http's Handler interface. To assist with some of the
 	utility behaviours of SilverStripe's controllers, BaseController can
 	be embedded in a controller type.
 *	DataObject - generally matches interface{}. See "DataObject Representation"
 	below.

## ORM

### DataObject Representation

Similar to the runtime's encoding/json package, goss attempts to be able to
map any DataObject read from the database into a corresponding struct, or into a map if there is no appropriate structure.

To have the ORM read into a predefined struct, two steps are required:

 1.	You need to register an empty instance of the type with the ORM,
	which maintains a mapping of ClassName values to such instances.
 2. In the struct, embed orm.DataObjectBase, which includes the common fields
	to all DataObjects.

Whenever an ORM query is made, the ORM will automatically store values from the database record into the struct's properties by name. Fields returned from the DB that are not in the struct are ignored.

If the ORM does not have a matching registered class, it will use DataObjectMap as the concrete type, which is a map of strings to interface{}.

A consequence of this is that it's possible to get a list of objects that contains both map and struct-based object.

(There is likely to be more work here.)

### DataList

### Configuration

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

'goss' only understands having one theme directory. It understands the templates, templates/Layout and templates/Include structure within this, but will expect to find all templates in the same theme.

### Implemented

This section lists features of the SilverStripe templating language that have been implemented. Some of the implementations may vary because of underlying differences in the systems.

 *	variable substitutions: $foo, {$foo}
 *	function substitutions: $foo(args), {$foo(args)}, $foo(arg1, arg2)
 *	<% if cond %>...<% end_if%> , <% else %> variation
 *	<% loop expr %>...<% end_loop %>
 *	<% with expr %>...<% end_with %>
 *	<% include %>, but not optional binding syntax
 *	operators: ||, &&, ==, !=, >, >=, <, <=, not, = as synonym for ==
 *	numeric literals (no decimals)
 *	string literals (backslashes in strings not handled)
 *	chained variable and functions, eg $foo.bar, $foo().bar("abc")
 *	main templates and main/layout arrangements, $Layout
 *	comments
 *	<% base_tag %>
 *	requirements injection
 *	<% require ... %>
 *	<% cached %> blocks are parsed correctly and contents substituted, with
 	no actual caching (compiled templates are, however, cached)
 *	XML and RAW formatting options on fields (implemented differently from SS,
	directly in the template injection rather than field function.)
 
### Not implemented

The following are not implemented. They are listed in approximate priority order for implementaton.

 *	else_if
 *	\$var
 *	shortcode handling
 *	backslash handling in string literals
 *	<%t ... %>
 *	<% include %> allows for an optional binding syntax for the included
	template. This extra syntax is not implemented.
 *	requirements combining / optimizing
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
