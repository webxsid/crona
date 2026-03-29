package messageregistry

import (
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
)

type Handler[M tea.Model] func(M, tea.Msg) (tea.Model, tea.Cmd, bool)

type Router[M tea.Model] struct {
	handlers map[reflect.Type]Handler[M]
}

func New[M tea.Model]() *Router[M] {
	return &Router[M]{handlers: map[reflect.Type]Handler[M]{}}
}

func (r *Router[M]) Register(example tea.Msg, handler Handler[M]) {
	r.handlers[reflect.TypeOf(example)] = handler
}

func (r *Router[M]) Handle(model M, msg tea.Msg) (tea.Model, tea.Cmd) {
	if msg == nil {
		return model, nil
	}
	if handler, ok := r.handlers[reflect.TypeOf(msg)]; ok {
		next, cmd, handled := handler(model, msg)
		if handled {
			return next, cmd
		}
	}
	return model, nil
}
