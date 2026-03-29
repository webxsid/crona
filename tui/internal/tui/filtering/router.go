package filtering

type Router[P comparable, M any, R any] struct {
	handlers map[P]func(*M) R
}

func New[P comparable, M any, R any]() *Router[P, M, R] {
	return &Router[P, M, R]{handlers: map[P]func(*M) R{}}
}

func (r *Router[P, M, R]) Register(pane P, handler func(*M) R) {
	r.handlers[pane] = handler
}

func (r *Router[P, M, R]) Resolve(pane P, model *M) (R, bool) {
	handler, ok := r.handlers[pane]
	if !ok {
		var zero R
		return zero, false
	}
	return handler(model), true
}
