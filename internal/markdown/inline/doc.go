// Package inline parses inline Markdown constructs within a block-scoped
// source span.
//
// It transforms span-referenced inline source text into structured
// ast.Inline nodes. Inline parsing is invoked only for content already
// delimited by the block parser and does not participate in block-level
// structure recognition.
//
// Parsing proceeds in three steps:
//
//	Scan  – converts a source span into lexical tokens
//	Build – consumes the token stream into a mutable item list and delimiter stack
//	Lower – converts the resulting item structure into ast.Inline nodes
//
// The scanner is purely lexical: it identifies delimiter runs, brackets,
// angle markers, escapes, and plain text without assigning semantic meaning.
//
// Build performs the actual inline parsing. It walks the token stream once,
// initially treating recognized syntax as provisional text while recording
// delimiter and bracket metadata. As sufficient context becomes available,
// it rewrites regions of the working item list into structured forms such as
// emphasis, strong emphasis, links, images, code spans, autolinks, and
// inline HTML.
//
// Lower performs the final structural conversion from the working item
// representation to []ast.Inline.
//
// This design separates lexical segmentation from semantic resolution while
// preserving byte spans into the original source throughout the parse.
package inline
