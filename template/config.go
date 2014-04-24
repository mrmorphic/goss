package template

import (
	"errors"
	"fmt"
	"github.com/mrmorphic/goss"
)

// A config object for the template package, which makes it more obvious when we're talking about config properties.
var configuration struct {
	initialised     bool
	ssRoot          string
	themeName       string
	siteUrl         string
	defaultProtocol string

	// derived properties
	templatesPath string
	cssPath       string
	layoutsPath   string
	includesPath  string
}

// Register getConfig to be provided configuration from the app.
func init() {
	configuration.initialised = false

	fns := []func(goss.ConfigProvider) error{getConfig}
	goss.RegisterInit(fns)
}

// getConfig is invoked when configuration is provided by the application. We extract out of it what we want,
// validate, and put the results in the config struct.
func getConfig(c goss.ConfigProvider) error {
	fmt.Printf("template.getConfig got %s\n", c)
	base := c.AsString("goss.ssroot")
	fmt.Printf("base is %s\n", base)
	if base == "" {
		return errors.New("goss template rendering requires configuration property 'ssroot' is set")
	}
	if base[len(base)-1] != '/' {
		base += "/"
	}

	theme := c.AsString("goss.theme")
	if theme == "" {
		return errors.New("goss template rendering requires configuration property 'theme' is set")
	}

	siteUrl := c.AsString("goss.siteUrl") // optional

	defaultProtocol := c.AsString("goss.defaultProtocol")
	if defaultProtocol == "" {
		defaultProtocol = "http"
	}
	if defaultProtocol != "http" && defaultProtocol != "https" {
		return fmt.Errorf("goss only supports defaultProtocol of 'http' or 'https', and not '%s'", defaultProtocol)
	}

	configuration.initialised = true
	configuration.ssRoot = base
	configuration.themeName = theme
	configuration.siteUrl = siteUrl
	configuration.defaultProtocol = defaultProtocol

	configuration.templatesPath = configuration.ssRoot + "themes/" + configuration.themeName + "/templates/"
	configuration.cssPath = configuration.ssRoot + "themes/" + configuration.themeName + "/css/"
	configuration.layoutsPath = "Layout/"
	configuration.includesPath = "Includes/"

	return nil
}
