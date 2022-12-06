package router_test

import (
	"testing"

	"code.olapie.com/sugar/testx"

	"code.olapie.com/router"
)

func TestNormalize(t *testing.T) {
	testx.Equal(t, "hello", router.Normalize("hello//"))
	testx.Equal(t, "hello/{id}", router.Normalize("hello/{id}/"))
	testx.Equal(t, "hello/{id}", router.Normalize("//hello/{id}/"))
	testx.True(t, router.Normalize("//") == "")
}

func TestIsStaticPath(t *testing.T) {
	testx.False(t, router.IsStatic("{a}"))
	testx.True(t, router.IsStatic("ab"))
	testx.False(t, router.IsParam("/a"))
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
			testx.True(t, router.IsParam(v))
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
			testx.False(t, router.IsParam(v))
		}
	})
}
