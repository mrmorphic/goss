// url package contains utility routines for URL handling that are either not provided by the runtime
// or are specific to SilverStripe.
package url

import (
	"fmt"
	"github.com/mrmorphic/goss/data"
	"net/http"
)

// LinkingMode returns one of 3 values:
// - "link" if none of this page or its children are current
// - "section" if a child of this page is open
// - "current" this page is open
func LinkingMode(r *http.Request, obj interface{}) string {
	// get relative path to obj
	// get relative path of request URL
	// if same -> current
	// if request path is a subset of obj path
	// otherwise link
	requestRel := r.URL.Path
	pageRel := data.Eval(obj, "Link").(string)
	fmt.Printf("Linking mode: requestRel: '%s', pageRel: '%s'\n", requestRel, pageRel)
	if requestRel == pageRel {
		return "current"
	} else if len(requestRel) < len(pageRel) && requestRel == pageRel[0:len(requestRel)] {
		return "section"
	}
	return "link"
}

func LinkOrCurrent(r *http.Request, obj interface{}) string {
	return "link"

}

func LinkOrSection(r *http.Request, obj interface{}) string {
	return "link"

}
