package router

import (
	"code.olapie.com/sugar/v2/xtype"
	"container/list"
	"fmt"
	"os"
	"reflect"
	"regexp"
	"strings"

	"code.olapie.com/sugar/v2/conv"
)

type nodeType int

const (
	staticNode   nodeType = iota // /users
	paramNode                    // /users/{id}
	wildcardNode                 // /users/{id}/photos/*
)

func (n nodeType) String() string {
	switch n {
	case staticNode:
		return "staticNode"
	case paramNode:
		return "paramNode"
	case wildcardNode:
		return "wildcardNode"
	default:
		return ""
	}
}

func getNodeType(segment string) nodeType {
	switch {
	case IsStatic(segment):
		return staticNode
	case IsParam(segment):
		return paramNode
	case IsWildcard(segment):
		return wildcardNode
	default:
		panic(fmt.Sprintf("Invalid segment: %s", segment))
	}
}

type node[H any] struct {
	typ       nodeType
	path      string // E.g. /items/{id}
	segment   string // E.g. items or {id}
	paramName string // E.g. id
	handlers  *list.List
	children  []*node[H]

	Input       any
	Outputs     []any
	Description string
	Sensitive   bool

	Metadata any
}

func newNodeList[H any](path string, handlers []H) *node[H] {
	path = Normalize(path)
	segments := strings.Split(path, "/")
	var head, p *node[H]
	for i, s := range segments {
		path := strings.Join(segments[:i+1], "/")
		n := newNode[H](path, s)
		if p != nil {
			p.children = []*node[H]{n}
		} else {
			head = n
		}
		p = n
	}
	if p != nil && len(handlers) > 0 {
		p.handlers = list.New()
		for _, h := range handlers {
			p.handlers.PushBack(h)
		}
	}
	return head
}

func newNode[H any](path, segment string) *node[H] {
	if len(strings.Split(segment, "/")) > 1 {
		panic(fmt.Sprintf("Invalid segment: %s", segment))
	}
	n := &node[H]{
		typ:     getNodeType(segment),
		path:    path,
		segment: segment,
	}
	switch n.typ {
	case paramNode:
		n.paramName = segment[1 : len(segment)-1]
	case wildcardNode:
		n.segment = segment[1:]
	default:
		break
	}
	return n
}

func newEmptyNode[H any]() *node[H] {
	return &node[H]{
		typ: staticNode,
	}
}

func (n *node[H]) Type() nodeType {
	return n.typ
}

func (n *node[H]) Path() string {
	return n.path
}

func (n *node[H]) IsEndpoint() bool {
	return n.handlers != nil && n.handlers.Len() > 0
}

func (n *node[H]) ListEndpoints() []*node[H] {
	var l []*node[H]
	if n.IsEndpoint() {
		l = append(l, n)
	}

	for _, child := range n.children {
		l = append(l, child.ListEndpoints()...)
	}
	return l
}

func (n *node[H]) Handler() *HandlerWrapper[H] {
	return (*HandlerWrapper[H])(n.handlers.Front())
}

func (n *node[H]) SetHandlers(handlers []H) {
	if n.handlers != nil && n.handlers.Len() > 0 {
		panic("Cannot overwrite handlers")
	}
	n.handlers = list.New()
	for _, h := range handlers {
		n.handlers.PushBack(h)
	}
}

func (n *node[H]) InsertPreHandlers(handlers []H) {
	for i := len(handlers) - 1; i >= 0; i-- {
		h := handlers[i]
		hStr := fmt.Sprint(h)
		found := false
		for e := n.handlers.Front(); e != nil; e = e.Next() {
			if any(h) == e.Value || hStr == fmt.Sprint(e.Value) {
				found = true
				break
			}
		}

		if found {
			handlers = append(handlers[:i], handlers[i+1:]...)
		}
	}

	if len(handlers) == 0 {
		return
	}

	hl := conv.ToList(handlers)
	n.handlers.PushFrontList(hl)
}

func (n *node[H]) Conflict(nod *node[H]) *xtype.Pair[*node[H]] {
	if n.typ != nod.typ {
		return nil
	}

	switch n.typ {
	case staticNode:
		if n.segment != nod.segment {
			return nil
		}

		if n.IsEndpoint() && nod.IsEndpoint() {
			return &xtype.Pair[*node[H]]{
				First:  n,
				Second: nod,
			}
		}
	case paramNode:
		if n.IsEndpoint() && nod.IsEndpoint() {
			return &xtype.Pair[*node[H]]{
				First:  n,
				Second: nod,
			}
		}
	case wildcardNode:
		return &xtype.Pair[*node[H]]{
			First:  n,
			Second: nod,
		}
	}

	for _, a := range n.children {
		for _, b := range nod.children {
			if v := a.Conflict(b); v != nil {
				return v
			}
		}
	}
	return nil
}

func (n *node[H]) Add(nod *node[H]) {
	var match *node[H]
	for _, child := range n.children {
		if v := child.Conflict(nod); v != nil {
			panic(fmt.Sprintf("Conflict: %s, %s", v.First.path, v.Second.path))
		}

		if child.segment == nod.segment {
			match = child
			break
		}
	}

	// Match: reuse the same node and append new nodes
	if match != nil {
		if len(nod.children) == 0 {
			match.handlers = nod.handlers
			return
		}

		for _, child := range nod.children {
			match.Add(child)
		}
		return
	}

	// Mismatch: append new nodes
	switch nod.typ {
	case staticNode:
		n.children = append([]*node[H]{nod}, n.children...)
	case paramNode:
		i := len(n.children) - 1
		for i >= 0 {
			if n.children[i].typ != wildcardNode {
				break
			}
			i--
		}

		if i < 0 {
			n.children = append([]*node[H]{nod}, n.children...)
		} else if i == len(n.children)-1 {
			n.children = append(n.children, nod)
		} else {
			n.children = append(n.children, nod)
			copy(n.children[i+2:], n.children[i+1:])
			n.children[i+1] = nod
		}
	case wildcardNode:
		n.children = append(n.children, nod)
	default:
		panic(fmt.Sprintf("Invalid node type: %v", nod.typ))
	}
}

func (n *node[H]) MatchPath(path string) (*node[H], map[string]string) {
	segments := strings.Split(path, "/")
	if segments[0] != "" {
		segments = append([]string{""}, segments...)
	}
	return n.Match(segments...)
}

func (n *node[H]) Match(segments ...string) (*node[H], map[string]string) {
	if len(segments) == 0 {
		if n.typ == wildcardNode {
			return n, nil
		}
		return nil, nil
	}

	first := segments[0]
	switch n.typ {
	case staticNode:
		if n.segment != first {
			return nil, nil
		}
		if len(segments) == 1 {
			if n.IsEndpoint() {
				return n, nil
			}
			// Perhaps some child nodes are wildcard node which can match empty node
			for _, child := range n.children {
				if child.typ == wildcardNode {
					return child, nil
				}
			}
			return nil, nil
		}
		if segments[1] == "" && n.IsEndpoint() {
			return n, nil
		}
		for _, child := range n.children {
			match, params := child.Match(segments[1:]...)
			if match != nil {
				return match, params
			}
		}
	case paramNode:
		var match *node[H]
		var params map[string]string
		if len(segments) == 1 || (segments[1] == "" && n.IsEndpoint()) {
			match = n
		} else {
			for _, child := range n.children {
				match, params = child.Match(segments[1:]...)
				if match != nil {
					break
				}
			}
		}

		if match != nil && match.IsEndpoint() {
			if params == nil {
				params = map[string]string{}
			}
			params[n.paramName] = first
			return match, params
		}
	case wildcardNode:
		if n.IsEndpoint() {
			return n, nil
		}
	}
	return nil, nil
}

func (n *node[H]) HandlerPath() string {
	reg := regexp.MustCompile(`\(\*(\w+)\)`)
	s := new(strings.Builder)
	for e := n.handlers.Front(); e != nil; e = e.Next() {
		if s.Len() > 0 {
			s.WriteString(", ")
		}

		var name string
		if s, ok := e.Value.(fmt.Stringer); ok {
			name = s.String()
		} else {
			name = reflect.TypeOf(e.Value).Name()
		}

		if strings.HasSuffix(name, "-fm") {
			name = name[:len(name)-3]
		}
		name = reg.ReplaceAllString(name, "$1")
		s.WriteString(shortPath(name))
	}
	return s.String()
}

var packagePath = func() string {
	s := os.Getenv("GOPATH")
	s = strings.TrimSpace(s)
	if len(s) == 0 {
		return ""
	}
	s = strings.TrimSuffix(s, "/")
	//Log.Println("GOPATH:", s)
	return s + "/src/"
}()

const goSrc = "/go/src/"

func relativePath(path string) string {
	if len(packagePath) > 0 {
		return strings.TrimPrefix(path, packagePath)
	}
	start := strings.Index(path, goSrc)
	if start > 0 {
		start += len(goSrc)
		path = path[start:]
	}
	return path
}

func shortPath(path string) string {
	path = relativePath(path)
	names := strings.Split(path, "/")
	for i := 0; i < len(names)-1; i++ {
		if len(names[i]) > 0 {
			names[i] = names[i][0:1]
		}
	}
	return strings.Join(names, "/")
}
