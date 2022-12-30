package router

import (
	"code.olapie.com/sugar/v2/xtest"
	"testing"
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

	pair := root.Conflict(newNodeList("/hello/world/{param}", hl))
	xtest.True(t, pair != nil)

	pair = root.Conflict(newNodeList("/hello/{world}", hl))
	xtest.True(t, pair == nil)

	pair = root.Conflict(newNodeList("/hello/{world}/{param}", hl))
	xtest.True(t, pair == nil)

	pair = root.Conflict(newNodeList("/hello/world/*", hl))
	xtest.True(t, pair == nil)
}
