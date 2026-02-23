// Package ast defines the semantic Markdown abstract syntax tree.
//
// The AST is the canonical representation of Markdown meaning in this
// system. It is renderer-facing and stable across parsing strategies.
// Types in this package must not contain raw Markdown source text,
// parsing artifacts, or dialect-specific behavior.
//
// All inline content is represented as structured Inline nodes.
// All block structure is explicit and fully formed.
package ast
