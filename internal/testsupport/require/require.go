package require

import (
	"errors"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func Equal[T any](t *testing.T, got, want T) {
	t.Helper()

	if !isEqual(got, want) {
		t.Fatalf("got: %v; want: %v", got, want)
	}
}

func NotEqual[T any](t *testing.T, got, want T) {
	t.Helper()

	if isEqual(got, want) {
		t.Fatalf("got: %v; expected values to be different", got)
	}
}

func True(t *testing.T, got bool) {
	t.Helper()

	if !got {
		t.Fatalf("got: false; want: true")
	}
}

func False(t *testing.T, got bool) {
	t.Helper()

	if got {
		t.Fatalf("got: true; want: false")
	}
}

func Nil(t *testing.T, got any) {
	t.Helper()

	if !isNil(got) {
		t.Fatalf("got: %v; want: nil", got)
	}
}

func NotNil(t *testing.T, got any) {
	t.Helper()

	if isNil(got) {
		t.Fatalf("got: nil; want: non-nil")
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func ErrorIs(t *testing.T, got, want error) {
	t.Helper()

	if !errors.Is(got, want) {
		t.Fatalf("got: %v; want: %v", got, want)
	}
}

func ErrorAs(t *testing.T, got error, target any) {
	t.Helper()

	if got == nil {
		t.Fatalf("got: nil; want assignable to: %T", target)
		return
	}

	if !errors.As(got, target) {
		t.Fatalf("got: %v; want assignable to: %T", got, target)
	}
}

func Contains(t *testing.T, got, substr string) {
	t.Helper()

	if !strings.Contains(got, substr) {
		t.Fatalf("got: %q; expected to contain: %q", got, substr)
	}
}

func MatchesRegexp(t *testing.T, got, pattern string) {
	t.Helper()

	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("invalid regexp %q: %v", pattern, err)
	}
	if !re.MatchString(got) {
		t.Fatalf("got: %q; want to match: %q", got, pattern)
	}
}

// Panics asserts that fn panics (with any value).
func Panics(t *testing.T, fn func()) {
	t.Helper()

	defer func() {
		if got := recover(); got == nil {
			t.Fatalf("expected panic, got none")
		}
	}()

	fn()
}

// PanicsError asserts that fn panics and that the panic value implements error.
// It returns the recovered error for further assertions.
func PanicsError(t *testing.T, fn func()) (err error) {
	t.Helper()

	defer func() {
		got := recover()
		if got == nil {
			t.Fatalf("expected panic, got none")
			return
		}

		e, ok := got.(error)
		if !ok {
			t.Fatalf("expected panic value to be error, got %T (%v)", got, got)
			return
		}

		err = e
	}()

	fn()
	return nil
}

// PanicsErrorContains asserts that fn panics with an error whose Error() message
// contains wantSubstring.
func PanicsErrorContains(t *testing.T, fn func(), wantSubstring string) {
	t.Helper()

	err := PanicsError(t, fn)
	if err == nil {
		t.Fatalf("expected non-nil error from PanicsError")
		return
	}

	if !strings.Contains(err.Error(), wantSubstring) {
		t.Fatalf("expected panic message to contain %q, got %q", wantSubstring, err.Error())
	}
}

type equaler[T any] interface {
	Equal(T) bool
}

// isEqual prefers a type's Equal method when present; otherwise it falls back to DeepEqual.
func isEqual[T any](got, want T) bool {
	if isNil(got) && isNil(want) {
		return true
	}

	if eq, ok := any(got).(equaler[T]); ok {
		return eq.Equal(want)
	}
	if eq, ok := any(want).(equaler[T]); ok {
		return eq.Equal(got)
	}

	return reflect.DeepEqual(got, want)
}

// isNil first checks for nil equality, then uses reflection to check typed nil inside an interface.
func isNil(v any) bool {
	if v == nil {
		return true
	}

	rv := reflect.ValueOf(v)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return rv.IsNil()
	}

	return false
}
