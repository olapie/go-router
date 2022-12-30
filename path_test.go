package router_test

import (
	"code.olapie.com/sugar/v2/xtest"
	"testing"

	"code.olapie.com/router"
)

func TestNormalize(t *testing.T) {
	xtest.Equal(t, "hello", router.Normalize("hello//"))
	xtest.Equal(t, "hello/{id}", router.Normalize("hello/{id}/"))
	xtest.Equal(t, "hello/{id}", router.Normalize("//hello/{id}/"))
	xtest.True(t, router.Normalize("//") == "")
}

func TestIsStaticPath(t *testing.T) {
	xtest.False(t, router.IsStatic("{a}"))
	xtest.True(t, router.IsStatic("ab"))
	xtest.False(t, router.IsParam("/a"))
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
			xtest.True(t, router.IsParam(v))
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
			xtest.False(t, router.IsParam(v))
		}
	})
}
