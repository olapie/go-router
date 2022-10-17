package router

type Endpoint[H any] struct {
	Scope string
	node  *node[H]
}

func (e *Endpoint[H]) Path() string {
	return e.node.path
}

func (e *Endpoint[H]) SetDescription(s string) *Endpoint[H] {
	e.node.Description = s
	return e
}

func (e *Endpoint[H]) Description() string {
	return e.node.Description
}

func (e *Endpoint[H]) HandlerPath() string {
	return e.node.HandlerPath()
}

func (e *Endpoint[H]) Handler() *HandlerWrapper[H] {
	return e.node.Handler()
}

func (e *Endpoint[H]) Input() any {
	return e.node.Input
}

func (e *Endpoint[H]) SetInput(m any) *Endpoint[H] {
	e.node.Input = m
	return e
}

func (e *Endpoint[H]) Sensitive() bool {
	return e.node.Sensitive
}

func (e *Endpoint[H]) SetSensitive(b bool) {
	e.node.Sensitive = b
}

func (e *Endpoint[H]) Metadata() any {
	return e.node.Metadata
}

func (e *Endpoint[H]) SetMetadata(m any) {
	e.node.Metadata = m
}
