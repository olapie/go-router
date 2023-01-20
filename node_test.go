package router

import (
	"testing"

	"code.olapie.com/sugar/v2/xtest"
)

func TestNewNode(t *testing.T) {
	n := newNode[struct{}]("*", "*")
	xtest.Equal(t, wildcardNode, n.typ)
	n = newNode[struct{}]("*file", "*file")
	xtest.Equal(t, wildcardNode, n.typ)

	n = newNode[struct{}]("{a}", "{a}")
	xtest.Equal(t, paramNode, n.typ)
	xtest.Equal(t, "a", n.paramName)
}

func TestNode_Conflict(t *testing.T) {
	hl := []func(){func() {}}
	root := newNodeList("/hello/world/{param}", hl)

	good := []string{
		"/hello/{world}",
		"/hello/{world}/{param}",
		"/hello/world/*",
	}

	bad := []string{
		"/hello/world/{param}",
	}

	for _, test := range good {
		conflicts := root.Conflict(newNodeList(test, hl))
		if len(conflicts) != 0 {
			t.Fatalf("expected no conflit: %s", test)
		}
	}

	for _, test := range bad {
		conflicts := root.Conflict(newNodeList(test, hl))
		if len(conflicts) == 0 {
			t.Fatalf("expected conflit: %s", test)
		}
	}
}
