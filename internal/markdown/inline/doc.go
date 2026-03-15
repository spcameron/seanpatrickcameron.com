// Package inline parses inline Markdown constructs within a block of text.
//
// It transforms raw inline Markdown text into structured ast.Inline nodes,
// supporting emphasis, strong emphasis, and other inline elements provided
// by the active dialect.
//
// Inline parsing operates only on text already scoped to a block and is
// independent of block-level structure.
//
// The parser is implemented as a small pipeline:
//
//	Scan     – converts a source span into a stream of inline events
//	Gather   – builds a working item stream and delimiter records
//	Resolve  – pairs compatible delimiters and constructs inline nodes
//	Finalize – emits the resulting ast.Inline nodes
//
// This structure separates lexical recognition from delimiter resolution,
// following the general strategy described in the CommonMark specification.
package inline
