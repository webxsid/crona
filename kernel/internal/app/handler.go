package app

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"crona/kernel/internal/core"
	corecommands "crona/kernel/internal/core/commands"
	"crona/kernel/internal/events"
	"crona/kernel/internal/notify"
	"crona/kernel/internal/runtime"
	"crona/kernel/internal/updatecheck"
	"crona/shared/protocol"
	sharedtypes "crona/shared/types"
)

type Handler struct {
	startedAt string
	info      sharedtypes.KernelInfo
	pingDB    func(context.Context) error
	core      *core.Context
	bus       *events.Bus
	timer     *corecommands.TimerService
	shutdown  func()
	envMode   string
	paths     runtime.Paths
	updater   *updatecheck.Service
	alerts    *notify.Service
}

func NewHandler(startedAt string, info sharedtypes.KernelInfo, pingDB func(context.Context) error, coreCtx *core.Context, bus *events.Bus, shutdown func(), envMode string, paths runtime.Paths, updater *updatecheck.Service, alerts *notify.Service) *Handler {
	return &Handler{
		startedAt: startedAt,
		info:      info,
		pingDB:    pingDB,
		core:      coreCtx,
		bus:       bus,
		timer:     corecommands.GetTimerService(coreCtx),
		shutdown:  shutdown,
		envMode:   envMode,
		paths:     paths,
		updater:   updater,
		alerts:    alerts,
	}
}

func (h *Handler) Stream(ctx context.Context, req protocol.Request, writer *json.Encoder) error {
	if req.Method != protocol.MethodEventsSubscribe {
		return errors.New("unsupported stream method")
	}

	eventsCh := make(chan sharedtypes.KernelEvent, 64)
	unsubscribe := h.bus.Subscribe(func(event sharedtypes.KernelEvent) {
		select {
		case eventsCh <- event:
		default:
		}
	})
	defer unsubscribe()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case event := <-eventsCh:
			if err := writer.Encode(protocol.Event{
				Type:    event.Type,
				Payload: event.Payload,
			}); err != nil {
				return err
			}
		}
	}
}

func (h *Handler) Handle(ctx context.Context, req protocol.Request) protocol.Response {
	if resp, ok := h.handleKernelMethods(ctx, req); ok {
		return resp
	}
	if resp, ok := h.handleWorkMethods(ctx, req); ok {
		return resp
	}
	if resp, ok := h.handleRuntimeMethods(ctx, req); ok {
		return resp
	}
	return protocol.Response{
		ID: req.ID,
		Error: &protocol.Error{
			Code:    "not_implemented",
			Message: "kernel method not implemented yet",
		},
	}
}

func (h *Handler) handleNoParams(req protocol.Request, fn func() (any, error)) protocol.Response {
	value, err := fn()
	if err != nil {
		return errorResponse(req.ID, err)
	}
	return mustResult(req.ID, value)
}

func handle[T any](req protocol.Request, fn func(T) (any, error)) protocol.Response {
	var input T
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &input); err != nil {
			return errorResponse(req.ID, err)
		}
	}
	value, err := fn(input)
	if err != nil {
		return errorResponse(req.ID, err)
	}
	return mustResult(req.ID, value)
}

func parseStartedAt(startedAt string) time.Time {
	t, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return time.Now().UTC()
	}
	return t
}

func mustResult(id string, value any) protocol.Response {
	body, err := json.Marshal(value)
	if err != nil {
		return protocol.Response{
			ID: id,
			Error: &protocol.Error{
				Code:    "internal_error",
				Message: err.Error(),
			},
		}
	}

	return protocol.Response{
		ID:     id,
		Result: body,
	}
}

func errorResponse(id string, err error) protocol.Response {
	code := "request_failed"
	var data json.RawMessage
	type codedError interface {
		ProtocolErrorCode() string
	}
	type dataError interface {
		ProtocolErrorData() any
	}
	if e, ok := err.(codedError); ok {
		if value := e.ProtocolErrorCode(); value != "" {
			code = value
		}
	}
	if e, ok := err.(dataError); ok {
		if payload := e.ProtocolErrorData(); payload != nil {
			body, marshalErr := json.Marshal(payload)
			if marshalErr == nil {
				data = body
			}
		}
	}
	return protocol.Response{
		ID: id,
		Error: &protocol.Error{
			Code:    code,
			Message: err.Error(),
			Data:    data,
		},
	}
}

func decodeObject(raw json.RawMessage) (map[string]json.RawMessage, error) {
	if len(raw) == 0 {
		return map[string]json.RawMessage{}, nil
	}
	var out map[string]json.RawMessage
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func decodeRequiredInt64(raw map[string]json.RawMessage, key string) (int64, error) {
	value, ok := raw[key]
	if !ok {
		return 0, errors.New(key + " is required")
	}
	var out int64
	if err := json.Unmarshal(value, &out); err != nil {
		return 0, err
	}
	return out, nil
}

func decodeOptionalStringFromMap(raw map[string]json.RawMessage, key string) (*string, bool, error) {
	value, ok := raw[key]
	if !ok {
		return nil, false, nil
	}
	if string(value) == "null" {
		return nil, true, nil
	}
	var out string
	if err := json.Unmarshal(value, &out); err != nil {
		return nil, false, err
	}
	return &out, true, nil
}

func decodeOptionalIntFromMap(raw map[string]json.RawMessage, key string) (*int, bool, error) {
	value, ok := raw[key]
	if !ok {
		return nil, false, nil
	}
	if string(value) == "null" {
		return nil, true, nil
	}
	var out int
	if err := json.Unmarshal(value, &out); err != nil {
		return nil, false, err
	}
	return &out, true, nil
}

func decodeOptionalInt64FromMap(raw map[string]json.RawMessage, key string) (*int64, bool, error) {
	value, ok := raw[key]
	if !ok {
		return nil, false, nil
	}
	if string(value) == "null" {
		return nil, true, nil
	}
	var out int64
	if err := json.Unmarshal(value, &out); err != nil {
		return nil, false, err
	}
	return &out, true, nil
}

func decodeOptionalBoolFromMap(raw map[string]json.RawMessage, key string) (*bool, bool, error) {
	value, ok := raw[key]
	if !ok {
		return nil, false, nil
	}
	if string(value) == "null" {
		return nil, true, nil
	}
	var out bool
	if err := json.Unmarshal(value, &out); err != nil {
		return nil, false, err
	}
	return &out, true, nil
}

func ptrTo[T any](value T) *T {
	ptr := new(T)
	*ptr = value
	return ptr
}
