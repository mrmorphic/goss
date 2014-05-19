package control

import (
	"errors"
	// "fmt"
	"github.com/mrmorphic/goss/cache"
	// "github.com/mrmorphic/goss/data"
	"github.com/mrmorphic/goss/orm"
	"net/http"
	"strconv"
	"time"
)

// Base type for BaseController. Goss doesn't directly create this; it is a base for the application
// to extend.
type BaseController struct {
	request *http.Request
}

func (ctl *BaseController) Init(r *http.Request) {
	ctl.request = r
}

func (ctl *BaseController) Menu(level int) (orm.DataList, error) {
	key := "goss_Menu_" + strconv.Itoa(level)
	result := cache.Get(key)
	if result != nil {
		return result.(orm.DataList), nil
	}

	if level == 1 {
		q := orm.NewQuery("SiteTree").Where("\"SiteTree_Live\".\"ParentID\"=0").Where("\"ShowInMenus\"=1").Sort("\"Sort\" ASC")
		v, e := q.Run()
		if e != nil {
			return nil, e
		}

		cache.Store(key, v, time.Minute)

		return v.(orm.DataList), nil
	}

	return nil, nil
}

func (ctl *BaseController) Level(level int) orm.DataObject {
	return nil

}

// Return the SiteConfig DataObject.
func (ctl *BaseController) SiteConfig() (obj interface{}, e error) {
	v := cache.Get("SiteConfig")
	if v != nil {
		return v, nil
	}

	q := orm.NewQuery("SiteConfig").Limit(0, 1)
	res, e := q.Run()
	if e != nil {
		return nil, e
	}

	items, _ := res.(orm.DataList).Items()
	if len(items) < 1 {
		return nil, errors.New("There is no SiteConfig record")
	}

	cache.Store("SiteConfig", items[0], 60*time.Second)

	return items[0], nil
}

// If the user is currently logged in, return a Member data object that represents the user. If logged out, return nil.
func (ctl *BaseController) CurrentMember() (obj *orm.DataObject, e error) {
	// @todo implement BaseController.CurrentMember
	return nil, nil
}
