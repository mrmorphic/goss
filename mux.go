package goss

import (
	"net/http"
	"regexp"
)

type muxRule struct {
	pattern    string
	patternReg *regexp.Regexp
	handler    http.HandlerFunc
}

var rules []*muxRule

// @todo consider making pattern a test. If a sting => regex. If a function with bool parameter, use that. If bool, use that.
func AddMuxRule(pattern string, handler http.HandlerFunc) {
	r := &muxRule{pattern: pattern, handler: handler}
	r.patternReg = regexp.MustCompile(pattern)
	rules = append(rules, r)
}

// When we get a request, we match the URL against the regexs in the rules until we find a match, and execute
// the handler.
func MuxServe(w http.ResponseWriter, r *http.Request) {
	for _, rule := range rules {
		if rule.patternReg.MatchString(r.URL.Path) {
			rule.handler(w, r)
			return
		}
	}

	// if we get here, there is no matching handler, so its a 404
	http.Error(w, "No handler for this URL", http.StatusNotFound)
}
