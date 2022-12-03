package router

import (
	"container/list"
	"context"

	"code.olapie.com/errors"
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

type HandlerWithError[IN any, OUT any] interface {
	Handle(ctx context.Context, in IN) (OUT, error)
}

type contextKey int

const keyNextHandler contextKey = iota

func Next[IN any, OUT any](ctx context.Context, in IN) (out OUT) {
	h, ok := ctx.Value(keyNextHandler).(*HandlerWrapper[Handler[IN, OUT]])
	if !ok {
		panic("cannot find next handler")
	}

	// if h.Next() is nil, then next call of NextWithError will stop
	ctx = context.WithValue(ctx, keyNextHandler, h.Next())
	return h.Handler().Handle(ctx, in)
}

func NextWithError[IN any, OUT any](ctx context.Context, in IN) (out OUT, err error) {
	h, ok := ctx.Value(keyNextHandler).(*HandlerWrapper[HandlerWithError[IN, OUT]])
	if !ok {
		err = errors.NotImplemented("cannot find next handler")
		return
	}

	// if h.Next() is nil, then next call of NextWithError will stop
	ctx = context.WithValue(ctx, keyNextHandler, h.Next())
	return h.Handler().Handle(ctx, in)
}

func WithNextHandler[H any](ctx context.Context, next *HandlerWrapper[H]) context.Context {
	return context.WithValue(ctx, keyNextHandler, next)
}
