// Package ir defines the intermediate representation produced by the
// Markdown block parser.
//
// The IR captures block structure and raw inline text but does not
// represent full semantic meaning. It exists as a parse-facing structure
// and is lowered into the semantic AST before rendering.
//
// IR types may contain raw source text and other parsing metadata.
// They are not stable and should not be consumed outside the front-end.
package ir
