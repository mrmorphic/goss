package control

import (
	"errors"
	"fmt"
	"github.com/mrmorphic/goss/orm"
	"net/http"
	"strconv"
)

// Base type for BaseController. Goss doesn't directly create this; it is a base for the application
// to extend.
type BaseController struct {
	Object  orm.DataObject
	Request *http.Request
	Output  http.ResponseWriter
}

func (c *BaseController) Init(w http.ResponseWriter, r *http.Request, object orm.DataObject) {
	c.Object = object
	c.Request = r
	c.Output = w
}

func (ctl *BaseController) Menu(level int) (set *orm.DataList, e error) {
	q := orm.NewQuery("SiteTree").Where("\"SiteTree_Live\".\"ParentID\"=0").Where("\"ShowInMenus\"=1").OrderBy("\"Sort\" ASC")
	set, e = q.Exec()
	return
}

// Return the SiteConfig DataObject.
func (ctl *BaseController) SiteConfig() (obj orm.DataObject, e error) {
	q := orm.NewQuery("SiteConfig").Limit(0, 1)
	res, e := q.Exec()
	if e != nil {
		return nil, e
	}

	if len(res.Items) < 1 {
		return nil, errors.New("There is no SiteConfig record")
	}

	return res.Items[0], nil
}

// If the user is currently logged in, return a Member data object that represents the user. If logged out, return nil.
func (ctl *BaseController) CurrentMember() (obj *orm.DataObject, e error) {
	// @todo implement BaseController.CurrentMember
	return nil, nil
}

// Given a data object, return a path to it. For a site object, this is the URLSegments of the
// descendents and the object. For other (non-hierarchical objects) this is an empty string, because
// there is no structure. Note this is not a complete Link, because the link is a function
// of the presentation layer, not the model. But this is helpful especially for SiteTree objects.
func (ctl *BaseController) Path(obj orm.DataObject, field string) (string, error) {
	fmt.Printf("BaseController::Path field %s in %s\n", field, obj)
	if obj == nil {
		return "", errors.New("Cannot determine path to empty object")
	}
	hier := orm.IsHierarchical(obj.GetStr("ClassName"))
	if !hier {
		fmt.Printf("BaseController::Path: not hierarchical\n")

		return "", nil
	}

	//	i,_ := obj.AsInt("ParentID")
	//	fmt.Printf("ParentID for object is %d\n", i)
	//fmt.Printf("Object is %s\n", obj)
	res := obj.GetStr(field)
	for parentID, e := obj.GetInt("ParentID"); e != nil && parentID > 0; {
		// @todo don't hardcode "SiteTree", derive the base class using metadata.
		q := orm.NewQuery("SiteTree").Where("\"SiteTree_Live\".\"ID\"=" + strconv.Itoa(parentID))
		ds, e := q.Exec()
		if e != nil {
			return "", e
		}
		obj = ds.First()
		res = obj.GetStr(field) + "/" + res
	}
	fmt.Printf("BaseController::Path: returning %s\n", res)

	return res, nil
}
