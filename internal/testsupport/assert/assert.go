package assert

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
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func NotEqual[T any](t *testing.T, got, want T) {
	t.Helper()

	if isEqual(got, want) {
		t.Errorf("got: %v; expected values to be different", got)
	}
}

func True(t *testing.T, got bool) {
	t.Helper()

	if !got {
		t.Errorf("got: false; want: true")
	}
}

func False(t *testing.T, got bool) {
	t.Helper()

	if got {
		t.Errorf("got: true; want: false")
	}
}

func Nil(t *testing.T, got any) {
	t.Helper()

	if !isNil(got) {
		t.Errorf("got: %v; want: nil", got)
	}
}

func NotNil(t *testing.T, got any) {
	t.Helper()

	if isNil(got) {
		t.Errorf("got: nil; want: non-nil")
	}
}

func NoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func ErrorIs(t *testing.T, got, want error) {
	t.Helper()

	if !errors.Is(got, want) {
		t.Errorf("got: %v; want: %v", got, want)
	}
}

func ErrorAs(t *testing.T, got error, target any) {
	t.Helper()

	if got == nil {
		t.Errorf("got: nil; want assignable to: %T", target)
		return
	}

	if !errors.As(got, target) {
		t.Errorf("got: %v; want assignable to: %T", got, target)
	}
}

func Contains(t *testing.T, got, substr string) {
	t.Helper()

	if !strings.Contains(got, substr) {
		t.Errorf("got: %q; expected to contain: %q", got, substr)
	}
}

func MatchesRegexp(t *testing.T, got, pattern string) {
	t.Helper()

	re, err := regexp.Compile(pattern)
	if err != nil {
		t.Fatalf("invalid regexp %q: %v", pattern, err)
	}
	if !re.MatchString(got) {
		t.Errorf("got: %q; want to match: %q", got, pattern)
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
