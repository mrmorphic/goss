package template

type chunkKind int

const (
	CHUNK_LITERAL  chunkKind = iota // literal text to emit
	CHUNK_BASE_TAG                  // <% base_tag %> substitution
	CHUNK_INCLUDE                   // process an include file
)

// chunk represents a piece of a template. The type indicates which piece this.
type chunk interface {
	kind() chunkKind
}

type chunkLiteral struct {
	content string
}

func (c *chunkLiteral) kind() chunkKind {
	return CHUNK_LITERAL
}

func newChunkLiteral(literal string) chunk {
	return &chunkLiteral{content: literal}
}

type chunkBaseTag struct{}

func (c *chunkBaseTag) kind() chunkKind {
	return CHUNK_BASE_TAG
}

func newChunkBaseTag() chunk {
	return &chunkBaseTag{}
}

// A chucnk to process an include file. The chunk has a reference to the compiled template for that file.
type chunkInclude struct {
	compiled *compiledTemplate
	// @todo include the bindings
}

func (c *chunkInclude) kind() chunkKind {
	return CHUNK_INCLUDE
}

func newChunkInclude(c *compiledTemplate) chunk {
	return &chunkInclude{c}
}
