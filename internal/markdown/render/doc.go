// Package render transforms the semantic Markdown AST into HTML.
//
// Rendering operates solely on ast types and does not depend on parsing
// details or raw Markdown source. It walks the AST and produces HTML
// nodes or strings according to the selected dialect and rendering rules.
package render
