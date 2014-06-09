package control

import (
	"errors"
	"github.com/mrmorphic/goss"
)

var configuration struct {
	// TTL for menu(n) cache, in seconds. 0 means no caching.
	cacheMenuTTL int

	// TTL for site config, in seconds. 0 means no caching.
	cacheSiteConfigTTL int

	// TTL for SiteTree cache used for nav generation. 0 means no caching.
	cacheSiteTreeNavTTL int
}

func init() {
	fns := []func(goss.ConfigProvider) error{setConfig}
	goss.RegisterInit(fns)
}

func setConfig(conf goss.ConfigProvider) error {
	i := conf.Get("goss.cache.menuTTL")
	ii, ok := i.(float64)
	if ok {
		configuration.cacheMenuTTL = int(ii)
	} else {
		return errors.New("goss expects config property goss.database.menuTTL to be of type 'int'.")
	}

	i = conf.Get("goss.cache.siteConfigTTL")
	ii, ok = i.(float64)
	if ok {
		configuration.cacheSiteConfigTTL = int(ii)
	} else {
		return errors.New("goss expects config property goss.database.siteConfigTTL to be of type 'int'.")
	}

	i = conf.Get("goss.cache.siteTreeTTL")
	ii, ok = i.(float64)
	if ok {
		configuration.cacheSiteTreeNavTTL = int(ii)
	} else {
		return errors.New("goss expects config property goss.database.siteTreeTTL to be of type 'int'.")
	}

	return nil
}
