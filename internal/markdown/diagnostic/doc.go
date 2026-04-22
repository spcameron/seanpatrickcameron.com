// Package diagnostic provides a minimal representation for reporting
// source-level errors and warnings.
//
// A Diagnostic records a message, a byte span into the source, and a
// severity. Diagnostics can be formatted with source context via
// Diagnostic.Format.
//
// DiagnosticError wraps a Diagnostic to satisfy the error interface.
package diagnostic
