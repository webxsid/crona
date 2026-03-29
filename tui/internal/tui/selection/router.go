package selection

type Router[K comparable, M any, R any] struct {
	handlers map[K]func(M) R
}

func New[K comparable, M any, R any]() *Router[K, M, R] {
	return &Router[K, M, R]{handlers: map[K]func(M) R{}}
}

func (r *Router[K, M, R]) Register(key K, handler func(M) R) {
	r.handlers[key] = handler
}

func (r *Router[K, M, R]) Resolve(key K, model M) (R, bool) {
	handler, ok := r.handlers[key]
	if !ok {
		var zero R
		return zero, false
	}
	return handler(model), true
}
