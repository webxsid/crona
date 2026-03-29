package dialogruntime

type Router[A any, R any] struct {
	handlers map[string]func(A) R
}

func New[A any, R any]() *Router[A, R] {
	return &Router[A, R]{handlers: map[string]func(A) R{}}
}

func (r *Router[A, R]) Register(kind string, handler func(A) R) {
	r.handlers[kind] = handler
}

func (r *Router[A, R]) Resolve(kind string, action A) (R, bool) {
	handler, ok := r.handlers[kind]
	if !ok {
		var zero R
		return zero, false
	}
	return handler(action), true
}
