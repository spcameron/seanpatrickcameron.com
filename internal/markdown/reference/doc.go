// Package reference provides utilities for validating and normalizing
// reference labels used in link and image definitions.
//
// A label is a constrained inline string that may include escapes and
// whitespace, but excludes unescaped bracket characters. Labels are
// matched using a canonical form to ensure consistent resolution.
package reference
