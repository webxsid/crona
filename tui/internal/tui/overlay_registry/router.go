package overlayregistry

import tea "github.com/charmbracelet/bubbletea"

type Handler[M tea.Model] func(M, tea.KeyMsg) (tea.Model, tea.Cmd, bool)

type Router[M tea.Model] struct {
	handlers map[string]Handler[M]
}

func New[M tea.Model]() *Router[M] {
	return &Router[M]{handlers: map[string]Handler[M]{}}
}

func (r *Router[M]) Register(key string, handler Handler[M]) {
	r.handlers[key] = handler
}

func (r *Router[M]) Resolve(model M, msg tea.KeyMsg) (tea.Model, tea.Cmd, bool) {
	if handler, ok := r.handlers[msg.String()]; ok {
		return handler(model, msg)
	}
	return model, nil, false
}

func (r *Router[M]) Handle(model M, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	next, cmd, handled := r.Resolve(model, msg)
	if handled {
		return next, cmd
	}
	return model, nil
}
