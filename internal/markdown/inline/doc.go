// Package inline parses inline Markdown constructs within a block of text.
//
// It transforms raw inline Markdown text into structured ast.Inline nodes,
// handling emphasis, strong, code spans, links, and other inline elements
// supported by the active dialect.
//
// Inline parsing operates only on text already scoped to a block and is
// independent of block-level structure.
package inline
