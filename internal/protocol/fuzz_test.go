package protocol

import (
	"bufio"
	"bytes"
	"testing"
)

func FuzzReadNDJSON(f *testing.F) {
	f.Add([]byte("{\"v\":1,\"id\":\"1\",\"type\":\"HELLO\"}\n"))
	f.Add([]byte("{}\n"))
	f.Add([]byte("not json\n"))
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ReadNDJSON(bufio.NewReader(bytes.NewReader(data)))
	})
}

func FuzzReadFrame(f *testing.F) {
	var buf bytes.Buffer
	m, _ := NewMessage("1", TypeHeartbeat, nil)
	_ = WriteFrame(&buf, m, 0)
	f.Add(buf.Bytes())
	f.Add([]byte{0xff, 0xff, 0xff, 0xff})
	f.Add([]byte{0, 0, 0, 1, '{'})
	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = ReadFrame(bytes.NewReader(data), 1<<20)
	})
}
