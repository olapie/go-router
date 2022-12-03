package router

import (
	"container/list"
	"context"
)

type HandlerWrapper[H any] list.Element

func (w *HandlerWrapper[H]) Handler() H {
	return w.Value.(H)
}

func (w *HandlerWrapper[H]) Next() *HandlerWrapper[H] {
	e := (*list.Element)(w)
	return (*HandlerWrapper[H])(e.Next())
}

type Handler[IN any, OUT any] interface {
	Handle(ctx context.Context, in IN) OUT
}

type HandlerFunc[IN any, OUT any] func(ctx context.Context, in IN) OUT

func (f HandlerFunc[IN, OUT]) Handle(ctx context.Context, in IN) OUT {
	return f(ctx, in)
}

type HandlerWithError[IN any, OUT any] interface {
	Handle(ctx context.Context, in IN) (OUT, error)
}

type HandlerFuncWithError[IN any, OUT any] func(ctx context.Context, in IN) (OUT, error)

func (f HandlerFuncWithError[IN, OUT]) Handle(ctx context.Context, in IN) (OUT, error) {
	return f(ctx, in)
}

type contextKey int

const keyNextHandler contextKey = iota

// Next calls the next handler and returns the result
// It panics if next handler doesn't exist
func Next[IN any, OUT any](ctx context.Context, in IN) (out OUT) {
	h, ok := ctx.Value(keyNextHandler).(*HandlerWrapper[Handler[IN, OUT]])
	if !ok {
		f, ok := ctx.Value(keyNextHandler).(*HandlerWrapper[HandlerFunc[IN, OUT]])
		if !ok {
			panic("cannot find next handler")
		}
		h = (*HandlerWrapper[Handler[IN, OUT]])(f)
	}

	// if h.Next() is nil, then next call of NextWithError will stop
	ctx = context.WithValue(ctx, keyNextHandler, h.Next())
	return h.Handler().Handle(ctx, in)
}

// NextWithError calls the next handler and returns the result
// It panics if next handler doesn't exist
func NextWithError[IN any, OUT any](ctx context.Context, in IN) (out OUT, err error) {
	h, ok := ctx.Value(keyNextHandler).(*HandlerWrapper[HandlerWithError[IN, OUT]])
	if !ok {
		f, ok := ctx.Value(keyNextHandler).(*HandlerWrapper[HandlerFuncWithError[IN, OUT]])
		if !ok {
			panic("cannot find next handler")
		}
		h = (*HandlerWrapper[HandlerWithError[IN, OUT]])(f)
	}

	// if h.Next() is nil, then next call of NextWithError will stop
	ctx = context.WithValue(ctx, keyNextHandler, h.Next())
	return h.Handler().Handle(ctx, in)
}

// WithNextHandler puts the wrapper of next handler into the context
// next must be returned from router
func WithNextHandler[H any](ctx context.Context, next *HandlerWrapper[H]) context.Context {
	return context.WithValue(ctx, keyNextHandler, next)
}
