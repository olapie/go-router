package router

import "container/list"

type HandlerWrapper[H any] list.Element

func (w *HandlerWrapper[H]) Handler() H {
	return w.Value.(H)
}

func (w *HandlerWrapper[H]) Next() *HandlerWrapper[H] {
	e := (*list.Element)(w)
	return (*HandlerWrapper[H])(e.Next())
}
