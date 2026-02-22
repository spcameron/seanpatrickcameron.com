// Package testkit provides helper constructors for building Markdown IR and AST
// values in tests.
//
// The helpers in this package exist solely to make test cases concise and readable.
// They are not intended for production use and do not represent a stable API.
//
// The constructors defined here may ignore invariants, omit validation, or change
// freely as the Markdown implementation evolves. Production code should construct
// IR via the block parser and AST via the build stage instead of using these helpers.
//
// This package lives under internal/markdown so that it is accessible to tests
// across markdown subpackages while remaining unavailable to external consumers.
package testkit
