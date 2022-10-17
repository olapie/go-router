package router

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewNode(t *testing.T) {
	n := newNode[struct{}]("*", "*")
	assert.Equal(t, wildcardNode, n.typ)
	n = newNode[struct{}]("*file", "*file")
	assert.Equal(t, wildcardNode, n.typ)

	n = newNode[struct{}]("{a}", "{a}")
	assert.Equal(t, paramNode, n.typ)
	assert.Equal(t, "a", n.paramName)
}

func TestNode_Conflict(t *testing.T) {
	hl := []func(){func() {}}
	root := newNodeList("/hello/world/{param}", hl)

	pair := root.Conflict(newNodeList("/hello/world/{param}", hl))
	assert.NotEmpty(t, pair)

	pair = root.Conflict(newNodeList("/hello/{world}", hl))
	assert.Empty(t, pair)

	pair = root.Conflict(newNodeList("/hello/{world}/{param}", hl))
	assert.Empty(t, pair)

	pair = root.Conflict(newNodeList("/hello/world/*", hl))
	assert.Empty(t, pair)
}
