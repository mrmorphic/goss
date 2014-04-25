package requirements

import (
	"bytes"
)

// DefaultRequirements is the default implementor of RequirementsProvider. It handles the basics of accepting requests
// to add CSS and JavaScript, and can inject the results in the correct places in generated HTML.
// @todo timestamp on files
// @todo combining
type DefaultRequirements struct {
	// list of javascript inclusions
	scripts []*inclusion

	// list of CSS inclusions
	css []*inclusion
}

// inclusion represents something to be included.
type inclusion struct {
	// if the inclusion is custom, this contains it
	custom string

	// if the inclusion if a path, this contains it
	path string

	// where to include, 'head' or 'body'
	where string

	// uniqueness value if provided
	uniqueness string
}

// Return the markup to inject for a CSS inclusion
func (i *inclusion) asCSS() string {
	if i.path != "" {
		return "\n" + `<link rel="stylesheet" type="text/css" href="` + i.path + `" />`
	}
	return `<style>
		` + i.custom + `
		</style>`
}

// Return the markup to inject for a javascript inclusion
func (i *inclusion) asJavascript() string {
	if i.path != "" {
		return `<script type="text/javascript" src="` + i.path + `"></script>`
	}
	return `<script type="javascript">
		` + i.custom + `
		</script>`
}

func NewRequirements() *DefaultRequirements {
	return &DefaultRequirements{}
}

// Javascript adds a path, assumed to be relative to web root of the SilverStripe application. It will be added
// before the </body>
func (r *DefaultRequirements) Javascript(path string) {
	inc := &inclusion{path: path, where: "body"}
	r.scripts = append(r.scripts, inc)
}

// CustomScript adds custom javascript code to the page. 'where' can be head or body.
func (r *DefaultRequirements) CustomScript(script string, where string, uniqueness string) {
	inc := &inclusion{custom: script, where: where, uniqueness: uniqueness}
	r.includeIfUnique(&r.scripts, inc)
}

// Add a CSS file to be included, relative to SS web root
func (r *DefaultRequirements) CSS(path string) {
	inc := &inclusion{path: path, where: "head"}
	r.css = append(r.css, inc)
}

// CustomCSS adds a CSS snippet inlined into the head of the document
func (r *DefaultRequirements) CustomCSS(css string, uniqueness string) {
	inc := &inclusion{custom: css, where: "head", uniqueness: uniqueness}
	r.includeIfUnique(&r.css, inc)
}

// add an inclusion to a list of inclusions, but only if the inclusion has a uniqueness
// value and it isn't already present in the requirements list.
func (r *DefaultRequirements) includeIfUnique(list *[]*inclusion, item *inclusion) {
	unique := true // true until proven otherwise
	for _, el := range *list {
		if el.uniqueness != "" && el.uniqueness == item.uniqueness {
			unique = false
		}
	}
	if unique {
		*list = append(*list, item)
	}
}

// InsertHeadTags injects any required markup before </head>. This will include
// CSS declarations in the order they were added, followed by javascript inclusions
// that were marked for inclusion in the head.
func (r *DefaultRequirements) InsertHeadTags(markup []byte) ([]byte, error) {
	// generate the markup to inject
	inject := ""
	// css first
	for _, i := range r.css {
		inject += string(i.asCSS())
	}
	// then js for the head
	for _, i := range r.scripts {
		if i.where == "head" {
			inject += i.asJavascript()
		}
	}

	i := bytes.Index(markup, []byte("</head>"))

	if i < 0 || inject == "" {
		return markup, nil
	}

	inject += "\n"
	var buf = &bytes.Buffer{}
	buf.Write(markup[0:i])
	buf.Write([]byte(inject))
	buf.Write(markup[i:])

	return buf.Bytes(), nil
}

// InsertBodyTags injects any requried markup before </body>. This will include javascript
// inclusions that were marked for inclusion in the body.
func (r *DefaultRequirements) InsertBodyTags(markup []byte) ([]byte, error) {
	// generate the markup to inject
	inject := ""

	// js for the body
	for _, i := range r.scripts {
		if i.where == "body" {
			inject += i.asJavascript()
		}
	}

	i := bytes.Index(markup, []byte("</body>"))

	if i < 0 || inject == "" {
		return markup, nil
	}

	var buf = bytes.NewBuffer(markup[0:i])
	buf.Write([]byte(inject))
	buf.Write(markup[i:])
	return buf.Bytes(), nil
}
