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
