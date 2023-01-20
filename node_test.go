package router

import (
	"testing"
)

func TestNewNode(t *testing.T) {
	n := newNode[struct{}]("*", "*")
	if n.typ != wildcardNode {
		t.Fatalf("node * type should be %s instead of %s", wildcardNode, n.typ)
	}
	n = newNode[struct{}]("*file", "*file")
	if n.typ != wildcardNode {
		t.Fatalf("node *file type should be %s instead of %s", wildcardNode, n.typ)
	}

	n = newNode[struct{}]("{a}", "{a}")
	if n.typ != paramNode {
		t.Fatalf("node type should be %s instead of %s", paramNode, n.typ)
	}
	if n.paramName != "a" {
		t.Fatalf("node param name should be \"a\" instead of %s", n.paramName)
	}
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
