package protocol

import (
	"encoding/json"
	"fmt"
)

// Transport-neutral IPC envelopes for the Unix socket migration.

type Request struct {
	ID     string          `json:"id"`
	Method string          `json:"method"`
	Params json.RawMessage `json:"params,omitempty"`
}

type Response struct {
	ID     string          `json:"id"`
	Result json.RawMessage `json:"result,omitempty"`
	Error  *Error          `json:"error,omitempty"`
}

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Error struct {
	Code    string          `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

const ErrorCodeStashConflict = "stash_conflict"

type RPCError struct {
	Code    string
	Message string
	Data    json.RawMessage
}

func (e *RPCError) Error() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *RPCError) DecodeData(out any) error {
	if e == nil || len(e.Data) == 0 || out == nil {
		return nil
	}
	return json.Unmarshal(e.Data, out)
}
