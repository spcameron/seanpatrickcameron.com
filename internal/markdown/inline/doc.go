// Package inline parses inline Markdown constructs within block-delimited
// source spans.
//
// It transforms inline source text into []ast.Inline while preserving byte
// spans into the original source. The package assumes block structure has
// already been determined and does not participate in block-level parsing.
//
// Parsing proceeds in stages:
//
//	Scan      – tokenizes inline source into lexical items
//	Build     – constructs and resolves a mutable item list plus delimiter stack
//	Finalize  – converts the resolved items into []ast.Inline
//
// The scanner is purely lexical. Semantic resolution happens during Build,
// where provisional text, delimiters, and brackets are rewritten into
// structured forms such as emphasis, links, images, code spans, autolinks,
// and inline HTML.
package inline
