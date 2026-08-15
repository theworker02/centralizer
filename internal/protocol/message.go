// Package protocol implements Centralizer Protocol 1.x.
//
// Messages are JSON objects with a type, correlation id, and payload.
// Stdio transports use NDJSON. Socket transports use a 4-byte big-endian
// length prefix followed by the same JSON body.
package protocol

import (
	"encoding/json"
	"fmt"

	"github.com/theworker02/centralizer/pkg/cir"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// Current version negotiated during HELLO.
const (
	Name    = "centralizer"
	Version = "1.0"
	Major   = 1
	Minor   = 0
)

// Type is a protocol message type.
type Type string

const (
	TypeHello         Type = "HELLO"
	TypeCapabilities  Type = "CAPABILITIES"
	TypeDescribe      Type = "DESCRIBE"
	TypeDescribeOK    Type = "DESCRIBE_OK"
	TypeCall          Type = "CALL"
	TypeResult        Type = "RESULT"
	TypeError         Type = "ERROR"
	TypeStreamOpen    Type = "STREAM_OPEN"
	TypeStreamData    Type = "STREAM_DATA"
	TypeStreamClose   Type = "STREAM_CLOSE"
	TypeHandleCreate  Type = "HANDLE_CREATE"
	TypeHandleRelease Type = "HANDLE_RELEASE"
	TypeGet           Type = "GET"
	TypeSet           Type = "SET"
	TypeHeartbeat     Type = "HEARTBEAT"
	TypeCancel        Type = "CANCEL"
	TypeShutdown      Type = "SHUTDOWN"
	TypeOK            Type = "OK"
)

// MaxFrameBytes is the default maximum accepted frame size (16 MiB).
const MaxFrameBytes = 16 << 20

// Message is a protocol envelope.
type Message struct {
	V       int             `json:"v"`
	ID      string          `json:"id"`
	Type    Type            `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// HelloPayload is sent by both sides during handshake.
type HelloPayload struct {
	Protocol string   `json:"protocol"`
	Name     string   `json:"name"`
	Features []string `json:"features,omitempty"`
	Runtime  string   `json:"runtime,omitempty"`
	Version  string   `json:"version,omitempty"`
}

// CallPayload invokes a function or method.
type CallPayload struct {
	Function string              `json:"function,omitempty"`
	Handle   string              `json:"handle,omitempty"`
	Method   string              `json:"method,omitempty"`
	TypeName string              `json:"type,omitempty"`
	Args     map[string]cir.Wire `json:"args,omitempty"`
}

// ResultPayload returns a CIR value.
type ResultPayload struct {
	Value cir.Wire `json:"value"`
}

// ErrorPayload is a structured protocol error.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Retry   bool   `json:"retry,omitempty"`
}

// StreamOpenPayload starts a stream.
type StreamOpenPayload struct {
	Name   string              `json:"name"`
	Args   map[string]cir.Wire `json:"args,omitempty"`
	Stream string              `json:"stream,omitempty"`
}

// StreamDataPayload carries one streamed value.
type StreamDataPayload struct {
	Stream string   `json:"stream"`
	Value  cir.Wire `json:"value"`
}

// StreamClosePayload ends a stream.
type StreamClosePayload struct {
	Stream string `json:"stream"`
	Error  string `json:"error,omitempty"`
}

// HandlePayload identifies a foreign object.
type HandlePayload struct {
	Handle   string              `json:"handle"`
	TypeName string              `json:"type,omitempty"`
	Property string              `json:"property,omitempty"`
	Value    *cir.Wire           `json:"value,omitempty"`
	Args     map[string]cir.Wire `json:"args,omitempty"`
}

// DescribePayload requests or returns a schema document (YAML or JSON text).
type DescribePayload struct {
	Schema string `json:"schema,omitempty"`
}

// EncodeArgs converts CIR args to wire args.
func EncodeArgs(args map[string]cir.Value) map[string]cir.Wire {
	if args == nil {
		return nil
	}
	out := make(map[string]cir.Wire, len(args))
	for k, v := range args {
		out[k] = v.ToWire()
	}
	return out
}

// DecodeArgs converts wire args to CIR args.
func DecodeArgs(args map[string]cir.Wire) (map[string]cir.Value, error) {
	if args == nil {
		return nil, nil
	}
	out := make(map[string]cir.Value, len(args))
	for k, w := range args {
		v, err := cir.FromWire(w)
		if err != nil {
			return nil, czerr.Wrap(czerr.ErrConversion, "arg "+k, err)
		}
		out[k] = v
	}
	return out, nil
}

// NewMessage builds an envelope with payload v encoded as JSON.
func NewMessage(id string, typ Type, payload any) (Message, error) {
	var raw json.RawMessage
	if payload != nil {
		b, err := json.Marshal(payload)
		if err != nil {
			return Message{}, err
		}
		raw = b
	}
	return Message{V: Major, ID: id, Type: typ, Payload: raw}, nil
}

// DecodePayload unmarshals m.Payload into dest.
func DecodePayload(m Message, dest any) error {
	if len(m.Payload) == 0 {
		return nil
	}
	if err := json.Unmarshal(m.Payload, dest); err != nil {
		return czerr.Wrap(czerr.ErrFrameInvalid, "payload", err)
	}
	return nil
}

// Compatible reports whether a peer protocol version can interoperate
// with this process. Major versions must match; minor may differ.
func Compatible(peer string) error {
	var maj, min int
	n, err := fmt.Sscanf(peer, "%d.%d", &maj, &min)
	if err != nil || n < 1 {
		return czerr.New(czerr.ErrProtocolMismatch, "unparseable protocol "+peer)
	}
	if maj != Major {
		return czerr.New(czerr.ErrProtocolMismatch, fmt.Sprintf("peer protocol %s incompatible with %s", peer, Version))
	}
	return nil
}
