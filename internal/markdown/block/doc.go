// Package block implements the Markdown front-end that transforms raw
// Markdown source into an intermediate representation (IR).
//
// It is responsible for block-level structure detection and container
// hierarchy (lists, blockquotes, code fences, etc.). Inline parsing is
// not performed at this stage; inline text remains raw in the IR.
//
// The parser does not produce semantic AST nodes directly.
package block
