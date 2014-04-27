package goss

//type NavigationProvider interface {
//	Menu(level int) DataList
//}

// ConfigProvider is an interface that can be used to get configuration values by a key. The keys form a hierarchy
// with "." separators.
type ConfigProvider interface {
	// Basic method for getting a configuration value.
	Get(key string) interface{}

	// Helper to return a configuration value as a string, which is the most common case.
	AsString(key string) string
}

// RequirementsProvider is an interface that can be used to collect required javascript and CSS, and inject it into
// output. goss/requirements package is the default implementation.
type RequirementsProvider interface {
	// @todo set/get combined_files_enabled
	// @todo set/get combined files folder
	// Add a javascript file to be included, relative to SS web root
	Javascript(path string)

	// Add custom javascript. 'script' is what to add. 'where' can be 'head' or 'body'.
	// uniqueness' prevents adding the same code added twice.
	CustomScript(script string, where string, uniqueness string)

	// Add a CSS file to be included, relative to SS web root
	CSS(path string)

	// Add custom CSS.
	CustomCSS(css string, uniqueness string)

	// Insert CSS (and custom script where 'top') just before </head>. If that tag can't be found,
	// returns an error.
	InsertHeadTags(markup []byte) ([]byte, error)

	// Insert javascript just befire </body>. If that tag can't be found, returns an error.
	InsertBodyTags(markup []byte) ([]byte, error)
}

// Evaluater represents something that can resolve a variable name to a value in it's context,
// or evaluate a named function with parameters in it's context. Evaluaters typically implement
// fallback behaviour, where if a name is not known in this context, it looks for a _fallback
// property or function, and if present, evaluates it and uses that for the original evaluation.
// The primary method is Get, but GetStr and GetInt are required as in many cases consumers will
// want a string or int.
type Evaluater interface {
	// Given a name and optional arguments, evaluate the name and return the value
	Get(name string, args ...interface{}) interface{}
	// Same as Get but converts to string
	GetStr(name string, args ...interface{}) string
	// Same as Get but converts to int
	GetInt(name string, args ...interface{}) (int, error)
}
