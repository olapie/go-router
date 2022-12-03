package router

import (
	"fmt"
	"net/url"
	"sort"
	"strings"

	"code.olapie.com/log"
)

// Router implements routing function
type Router[H any] struct {
	rootNodes map[string]*node[H]
	basePath  string
	handlers  []H

	children []*Router[H]
	nodes    []*node[H]
}

// New a Router
func New[H any]() *Router[H] {
	r := &Router[H]{
		rootNodes: make(map[string]*node[H], 4),
	}
	r.rootNodes[""] = newEmptyNode[H]()
	return r
}

func (r *Router[H]) clone() *Router[H] {
	nr := &Router[H]{
		rootNodes: r.rootNodes,
		basePath:  r.basePath,
	}
	r.children = append(r.children, nr)
	nr.handlers = make([]H, len(r.handlers))
	copy(nr.handlers, r.handlers)
	return nr
}

func (r *Router[H]) BasePath() string {
	return r.basePath
}

// SetGlobalHandlers inserts pre handlers for all nodes in this router and its child routers
func (r *Router[H]) InsertGlobalPreHandlers(handlers ...H) {
	for _, node := range r.nodes {
		if node.IsEndpoint() {
			node.InsertPreHandlers(handlers)
		}
	}

	for _, h := range handlers {
		if r.ContainsHandler(h) {
			panic(fmt.Sprintf("Duplicate handler: %v", h))
		}
	}
	r.handlers = append(handlers, r.handlers...)
}

// Group returns a new Router[H] whose basePath is r.basePath+path
func (r *Router[H]) Group(path string) *Router[H] {
	if path == "/" {
		log.S().Panic(`Not allowed to create group "/"`)
	}

	nr := r.clone()
	// support empty path
	if len(path) > 0 {
		nr.basePath = Normalize(r.basePath + "/" + path)
	}
	return nr
}

// Use returns a new Router[H] with global handlers which will be bound with all new path patterns
// This can be used to add interceptors
func (r *Router[H]) Use(handlers ...H) *Router[H] {
	nr := r.clone()
	for _, h := range handlers {
		name := fmt.Sprint(h)
		found := false
		for _, nh := range nr.handlers {
			if fmt.Sprint(nh) == name {
				found = true
				break
			}
		}

		if !found {
			nr.handlers = append(nr.handlers, h)
		}
	}
	return nr
}

// Match finds handlers and parses path parameters according to method and path
func (r *Router[H]) Match(scope string, path string) (*Endpoint[H], map[string]string) {
	segments := strings.Split(path, "/")
	if segments[0] != "" {
		segments = append([]string{""}, segments...)
	}

	root := r.rootNodes[scope]
	global := r.rootNodes[""]
	if root == nil {
		root = global
	}

	n, params := root.Match(segments...)
	if n == nil && root != global {
		n, params = global.Match(segments...)
	}

	if n == nil {
		return nil, map[string]string{}
	}

	unescaped := make(map[string]string, len(params))
	for k, v := range params {
		uv, err := url.PathUnescape(v)
		if err != nil {
			log.G().Error("unescape path param", log.String("param", v), log.Error(err))
			unescaped[k] = v
		} else {
			unescaped[k] = uv
		}
	}
	return &Endpoint[H]{
		Scope: scope,
		node:  n,
	}, unescaped
}

func (r *Router[H]) MatchScopes(path string) []string {
	var a []string
	for m := range r.rootNodes {
		if rt, _ := r.Match(m, path); rt != nil {
			a = append(a, m)
		}
	}
	return a
}

// Bind binds scope, path with handlers
func (r *Router[H]) Bind(scope, path string, handlers ...H) *Endpoint[H] {
	if path == "" {
		log.G().Panic("path is empty")
	}

	if len(handlers) == 0 {
		log.G().Panic("handlers cannot be empty")
	}

	scope = strings.ToUpper(scope)

	hl := make([]H, len(r.handlers)+len(handlers))
	copy(hl, r.handlers)
	copy(hl[len(r.handlers):], handlers)

	root := r.createRoot(scope)
	global := r.rootNodes[""]
	path = Normalize(r.basePath + "/" + path)
	if path == "" {
		if root.IsEndpoint() {
			panic(fmt.Sprintf("Conflict: %s, %s", scope, r.basePath))
		}

		if global.IsEndpoint() {
			panic(fmt.Sprintf("Conflict: %s", r.basePath))
		}
		root.SetHandlers(hl)
	} else {
		nl := newNodeList(path, hl)
		if pair := global.Conflict(nl); pair != nil {
			first := pair.First.Path()
			second := pair.Second.Path()
			panic(fmt.Sprintf("Conflict: %s, %s %s", first, scope, second))
		}
		root.Add(nl)
		r.nodes = append(r.nodes, nl)
	}
	n, _ := root.MatchPath(path)
	return &Endpoint[H]{
		Scope: scope,
		node:  n,
	}
}

func (r *Router[H]) createRoot(scope string) *node[H] {
	root := r.rootNodes[scope]
	if root == nil {
		root = newEmptyNode[H]()
		r.rootNodes[scope] = root
		r.nodes = append(r.nodes, root)
	}
	return root
}

// Print prints all path trees
func (r *Router[H]) Print() {
	for method, root := range r.rootNodes {
		nodes := root.ListEndpoints()
		for _, n := range nodes {
			log.S().Debugf("%-5s %s\t%s", method, n.Path(), n.HandlerPath())
		}
	}
}

func (r *Router[H]) ListRoutes() []*Endpoint[H] {
	l := make([]*Endpoint[H], 0, 10)
	for scope, root := range r.rootNodes {
		for _, e := range root.ListEndpoints() {
			l = append(l, &Endpoint[H]{
				Scope: scope,
				node:  e,
			})
		}
	}
	sort.Slice(l, func(i, j int) bool {
		return strings.Compare(l[i].node.path, l[j].node.path) < 0
	})
	return l
}

func (r *Router[H]) ContainsHandler(h H) bool {
	s := fmt.Sprint(h)
	for _, h := range r.handlers {
		if s == fmt.Sprint(h) {
			return true
		}
	}
	return false
}
