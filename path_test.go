package router

import (
	"testing"
)

func TestNormalize(t *testing.T) {
	tests := map[string]string{
		"hello//":       "hello",
		"hello/{id}/":   "hello/{id}",
		"//hello/{id}/": "hello/{id}",
		"//":            "",
	}

	for path, expected := range tests {
		if got := Normalize(path); got != expected {
			t.Fatalf("expected: %s, but got: %s", expected, got)
		}
	}
}

func TestIsStaticPath(t *testing.T) {
	good := []string{"ab", "a", "1"}
	bad := []string{"{a}", "{", "}", "/", "/a"}
	for _, v := range good {
		if !IsStatic(v) {
			t.Fatal(v)
		}
	}

	for _, v := range bad {
		if IsStatic(v) {
			t.Fatal(v)
		}
	}
}

func TestIsParamPath(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		trueCases := []string{
			"{a_b}",
			"{_b}",
			"{_1}",
			"{a1}",
			"{a1_}",
		}
		for _, v := range trueCases {
			if !IsParam(v) {
				t.Fatalf("expected param: %s", v)
			}
		}
	})
	t.Run("false", func(t *testing.T) {
		falseCases := []string{
			"{/a}",
			"c",
			"{_}",
			"{__}",
			"{a",
			"{1}",
			"{1_a}",
		}
		for _, v := range falseCases {
			if IsParam(v) {
				t.Fatalf("expected not param: %s", v)
			}
		}
	})
}
