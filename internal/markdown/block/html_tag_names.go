package block

// htmlBlockTagList enumerates HTML tag names that are recognized as
// block-level HTML elements for parsing purposes.
var htmlBlockTagList = []string{
	"address",
	"article",
	"aside",
	"blockquote",
	"body",
	"details",
	"dialog",
	"div",
	"dl",
	"fieldset",
	"figcaption",
	"figure",
	"footer",
	"form",
	"h1",
	"h2",
	"h3",
	"h4",
	"h5",
	"h6",
	"header",
	"hr",
	"html",
	"main",
	"menu",
	"nav",
	"ol",
	"p",
	"pre",
	"section",
	"table",
	"tbody",
	"td",
	"tfoot",
	"th",
	"thead",
	"tr",
	"ul",
}

// htmlBlockTags provides a set lookup for htmlBlockTagList.
var htmlBlockTags = buildHTMLBlockTagSet()

func buildHTMLBlockTagSet() map[string]struct{} {
	m := make(map[string]struct{}, len(htmlBlockTagList))
	for _, tag := range htmlBlockTagList {
		m[tag] = struct{}{}
	}

	return m
}
