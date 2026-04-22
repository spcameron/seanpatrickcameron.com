// Package ir defines the intermediate representation produced by the
// block parser.
//
// The IR captures block structure and inline source spans, but does not
// encode full inline semantics. It serves as a transient form between
// block parsing and AST construction.
//
// IR values are parse-facing and not intended for use outside the
// compilation pipeline.
package ir
