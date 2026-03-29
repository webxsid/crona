package keyregistry

import tea "github.com/charmbracelet/bubbletea"

type Handler[M tea.Model] func(M, tea.KeyMsg) (tea.Model, tea.Cmd, bool)

type Router[M tea.Model, V comparable, P comparable] struct {
	global map[string]Handler[M]
	views  map[V]map[string]Handler[M]
	panes  map[V]map[P]map[string]Handler[M]
}

func New[M tea.Model, V comparable, P comparable]() *Router[M, V, P] {
	return &Router[M, V, P]{
		global: map[string]Handler[M]{},
		views:  map[V]map[string]Handler[M]{},
		panes:  map[V]map[P]map[string]Handler[M]{},
	}
}

func (r *Router[M, V, P]) RegisterGlobal(key string, handler Handler[M]) {
	r.global[key] = handler
}

func (r *Router[M, V, P]) RegisterView(view V, key string, handler Handler[M]) {
	if _, ok := r.views[view]; !ok {
		r.views[view] = map[string]Handler[M]{}
	}
	r.views[view][key] = handler
}

func (r *Router[M, V, P]) RegisterPane(view V, pane P, key string, handler Handler[M]) {
	if _, ok := r.panes[view]; !ok {
		r.panes[view] = map[P]map[string]Handler[M]{}
	}
	if _, ok := r.panes[view][pane]; !ok {
		r.panes[view][pane] = map[string]Handler[M]{}
	}
	r.panes[view][pane][key] = handler
}

func (r *Router[M, V, P]) Handle(model M, view V, pane P, msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	key := msg.String()
	if panes, ok := r.panes[view]; ok {
		if handlers, ok := panes[pane]; ok {
			if handler, ok := handlers[key]; ok {
				next, cmd, handled := handler(model, msg)
				if handled {
					return next, cmd
				}
			}
		}
	}
	if handlers, ok := r.views[view]; ok {
		if handler, ok := handlers[key]; ok {
			next, cmd, handled := handler(model, msg)
			if handled {
				return next, cmd
			}
		}
	}
	if handler, ok := r.global[key]; ok {
		next, cmd, handled := handler(model, msg)
		if handled {
			return next, cmd
		}
	}
	return model, nil
}
