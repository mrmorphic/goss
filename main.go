package goss

// global configuration
var configuration ConfigProvider

var setupFunctions []func(ConfigProvider) error

// SetConfig tells goss the configuration object to use. This should be called before requests are accepted.
// The configuration properties that goss understands will be read at this point.
func SetConfig(conf ConfigProvider) error {
	configuration = conf
	for _, fn := range setupFunctions {
		e := fn(conf)
		if e != nil {
			return e
		}
	}

	return nil
}

// RegisterInit accepts a list of initialisation functions to be called whenever configuration is set.
func RegisterInit(funcs []func(ConfigProvider) error) {
	for _, f := range funcs {
		setupFunctions = append(setupFunctions, f)
	}
}
