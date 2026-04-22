// Package testkit provides helper constructors for building Markdown IR and AST
// values in tests.
//
// The helpers in this package exist solely to make test cases concise and readable.
// They are not intended for production use and do not represent a stable API.
//
// The constructors defined here may ignore invariants, omit validation, or change
// freely as the Markdown implementation evolves. Production code should construct
// IR via the block parser and AST via the build stage instead of using these helpers.
package testkit
