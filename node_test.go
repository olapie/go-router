package router

import (
	"testing"

	"code.olapie.com/sugar/testx"
)

func TestNewNode(t *testing.T) {
	n := newNode[struct{}]("*", "*")
	testx.Equal(t, wildcardNode, n.typ)
	n = newNode[struct{}]("*file", "*file")
	testx.Equal(t, wildcardNode, n.typ)

	n = newNode[struct{}]("{a}", "{a}")
	testx.Equal(t, paramNode, n.typ)
	testx.Equal(t, "a", n.paramName)
}

func TestNode_Conflict(t *testing.T) {
	hl := []func(){func() {}}
	root := newNodeList("/hello/world/{param}", hl)

	pair := root.Conflict(newNodeList("/hello/world/{param}", hl))
	testx.True(t, pair != nil)

	pair = root.Conflict(newNodeList("/hello/{world}", hl))
	testx.True(t, pair == nil)

	pair = root.Conflict(newNodeList("/hello/{world}/{param}", hl))
	testx.True(t, pair == nil)

	pair = root.Conflict(newNodeList("/hello/world/*", hl))
	testx.True(t, pair == nil)
}
