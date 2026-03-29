package eventregistry

import (
	tea "github.com/charmbracelet/bubbletea"
)

type Handler[M tea.Model, E any] func(M, E) (tea.Model, tea.Cmd, bool)

type Router[M tea.Model, E any] struct {
	handlers map[string]Handler[M, E]
}

func New[M tea.Model, E any]() *Router[M, E] {
	return &Router[M, E]{handlers: map[string]Handler[M, E]{}}
}

func (r *Router[M, E]) Register(eventType string, handler Handler[M, E]) {
	r.handlers[eventType] = handler
}

func (r *Router[M, E]) Handle(model M, eventType string, event E) (tea.Model, tea.Cmd) {
	if handler, ok := r.handlers[eventType]; ok {
		next, cmd, handled := handler(model, event)
		if handled {
			return next, cmd
		}
	}
	return model, nil
}
