package control

import (
	"github.com/mrmorphic/goss/data"
	"github.com/mrmorphic/goss/orm"
	"strconv"
)

// A utility type for embedding in models to provide a base set of functionality common to pages.
type DataObjectBase struct {
	ID         int
	ClassName  string
	ParentID   int // Only hierarchy has it, but put it here for convenience
	Title      string
	MenuTitle  string
	URLSegment string
}

// Return MenuTitle, or Title if MenuTitle is blank
func (d *DataObjectBase) GetMenuTitle() string {
	if d.MenuTitle == "" {
		return d.Title
	}
	return d.MenuTitle
}

// Generate a BaseHRef-relative link to this page
func (d *DataObjectBase) Link(args ...string) string {
	hier := orm.IsHierarchical(d.ClassName)
	if !hier {
		return ""
	}

	obj := interface{}(d)
	res := d.URLSegment

	// @todo data.Eval.(int) may fail for a map where the ParentID may have type of string
	for parentID := data.Eval(obj, "ParentID").(int); parentID > 0; {
		// @todo don't hardcode "SiteTree", derive the base class using metadata.
		q := orm.NewQuery("SiteTree").Where("\"SiteTree_Live\".\"ID\"=" + strconv.Itoa(parentID))
		ds, e := q.Run()
		if e != nil {
			return ""
		}
		items, _ := ds.(orm.DataList).Items()
		obj = items[0]
		res = data.Eval(obj, "URLSegment").(string) + "/" + res
	}

	for _, a := range args {
		res += "/" + a
	}

	if res == "home" {
		res = "/"
	}

	return res
}
