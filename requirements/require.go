package requirements

import (
	"bytes"
	// "fmt"
)

type DefaultRequirements struct {
	scripts []*inclusion
	css     []*inclusion
}

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

func (i *inclusion) asCSS() string {
	if i.path != "" {
		return `<link rel="stylesheet" type="text/css" href="` + i.path + `" />`
	}
	return `<style>
		` + i.custom + `
		</style>`
}

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

func (r *DefaultRequirements) Javascript(path string) {
	inc := &inclusion{path: path, where: "body"}
	r.scripts = append(r.scripts, inc)
}

func (r *DefaultRequirements) CustomScript(script string, where string, uniqueness string) {
	inc := &inclusion{custom: script, where: where, uniqueness: uniqueness}
	r.includeIfUnique(&r.scripts, inc)
}

// Add a CSS file to be included, relative to SS web root
func (r *DefaultRequirements) CSS(path string) {
	inc := &inclusion{path: path, where: "head"}
	r.css = append(r.css, inc)
}

func (r *DefaultRequirements) CustomCSS(css string, uniqueness string) {
	inc := &inclusion{custom: css, where: "head", uniqueness: uniqueness}
	r.includeIfUnique(&r.css, inc)
}

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

	var buf = bytes.NewBuffer(markup[0:i])
	buf.Write([]byte(inject))
	buf.Write(markup[i:])
	return buf.Bytes(), nil
}

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
