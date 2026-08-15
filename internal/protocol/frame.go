package protocol

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"io"

	"github.com/theworker02/centralizer/pkg/czerr"
)

// WriteNDJSON writes one message as a single JSON line.
func WriteNDJSON(w io.Writer, m Message) error {
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(m); err != nil {
		return czerr.Wrap(czerr.ErrTransportFailure, "write ndjson", err)
	}
	return nil
}

// ReadNDJSON reads one JSON line into a message.
func ReadNDJSON(r *bufio.Reader) (Message, error) {
	line, err := r.ReadBytes('\n')
	if err != nil {
		return Message{}, czerr.Wrap(czerr.ErrTransportFailure, "read ndjson", err)
	}
	if len(line) > MaxFrameBytes {
		return Message{}, czerr.ErrPayloadTooLarge
	}
	var m Message
	if err := json.Unmarshal(line, &m); err != nil {
		return Message{}, czerr.Wrap(czerr.ErrFrameInvalid, "ndjson", err)
	}
	if m.Type == "" {
		return Message{}, czerr.New(czerr.ErrFrameInvalid, "missing type")
	}
	return m, nil
}

// WriteFrame writes a length-prefixed JSON frame.
func WriteFrame(w io.Writer, m Message, max int) error {
	if max <= 0 {
		max = MaxFrameBytes
	}
	body, err := json.Marshal(m)
	if err != nil {
		return err
	}
	if len(body) > max {
		return czerr.ErrPayloadTooLarge
	}
	var hdr [4]byte
	binary.BigEndian.PutUint32(hdr[:], uint32(len(body)))
	if _, err := w.Write(hdr[:]); err != nil {
		return czerr.Wrap(czerr.ErrTransportFailure, "write header", err)
	}
	if _, err := w.Write(body); err != nil {
		return czerr.Wrap(czerr.ErrTransportFailure, "write body", err)
	}
	return nil
}

// ReadFrame reads a length-prefixed JSON frame.
func ReadFrame(r io.Reader, max int) (Message, error) {
	if max <= 0 {
		max = MaxFrameBytes
	}
	var hdr [4]byte
	if _, err := io.ReadFull(r, hdr[:]); err != nil {
		return Message{}, czerr.Wrap(czerr.ErrTransportFailure, "read header", err)
	}
	n := int(binary.BigEndian.Uint32(hdr[:]))
	if n < 0 || n > max {
		return Message{}, czerr.ErrPayloadTooLarge
	}
	body := make([]byte, n)
	if _, err := io.ReadFull(r, body); err != nil {
		return Message{}, czerr.Wrap(czerr.ErrTransportFailure, "read body", err)
	}
	var m Message
	if err := json.Unmarshal(body, &m); err != nil {
		return Message{}, czerr.Wrap(czerr.ErrFrameInvalid, "frame json", err)
	}
	if m.Type == "" {
		return Message{}, czerr.New(czerr.ErrFrameInvalid, "missing type")
	}
	return m, nil
}
