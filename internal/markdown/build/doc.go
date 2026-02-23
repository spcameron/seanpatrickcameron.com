// Package build constructs the semantic Markdown AST from block-level IR.
//
// It transforms the intermediate representation produced by the block
// parser into a fully-formed ast.Document. During this process, raw
// inline text is parsed into structured inline nodes.
//
// The resulting AST is complete and safe for rendering or analysis.
package build
