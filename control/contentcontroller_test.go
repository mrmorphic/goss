package control

import (
	"fmt"
	"github.com/mrmorphic/goss/orm"
	"testing"
)

func TestRequireJS(t *testing.T) {
	cc := &ContentController{}
	obj := orm.NewDataObjectMap()

	cc.Init(obj)

	// test that properties on ContentController can be fetched
	v := cc.Get("name")
	fmt.Printf("name from contentcontroller is %s\n", v)
	// test that properties on BaseController can be fetched
	// test that properties on object

	// source := `<html><head><title>x</title></head><body><% require javascript("themes/simple/javascript/test.js") %><div>test</div></body></html>`
	// context := map[string]interface{}{}

	// b, e := compileAndExecute(source, context)
	// if e != nil {
	// 	t.Error(e.Error())
	// 	return
	// }
}
