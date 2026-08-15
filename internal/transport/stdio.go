package transport

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"

	"github.com/theworker02/centralizer/internal/protocol"
	"github.com/theworker02/centralizer/internal/security"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// Stdio is a supervised child process speaking NDJSON on stdin/stdout.
type Stdio struct {
	Name string
	Argv []string
	Dir  string
	Env  []string

	mu     sync.Mutex
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Reader
	stderr lockedBuffer
	done   chan error
}

// lockedBuffer is a bytes.Buffer that is safe for concurrent Write and String.
// os/exec copies child stderr onto Write from another goroutine.
type lockedBuffer struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (l *lockedBuffer) Write(p []byte) (int, error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.Write(p)
}

func (l *lockedBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.b.String()
}

func (s *Stdio) Kind() Kind { return KindStdio }

func (s *Stdio) Open(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.Argv) == 0 {
		return czerr.New(czerr.ErrTransportFailure, "empty argv")
	}
	cmd := exec.CommandContext(ctx, s.Argv[0], s.Argv[1:]...)
	cmd.Dir = s.Dir
	if s.Env != nil {
		cmd.Env = security.FilterEnv(s.Env)
	} else {
		cmd.Env = security.FilterEnv(os.Environ())
	}
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return czerr.Wrap(czerr.ErrTransportFailure, "stdin", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return czerr.Wrap(czerr.ErrTransportFailure, "stdout", err)
	}
	cmd.Stderr = &s.stderr
	if err := cmd.Start(); err != nil {
		return czerr.Wrap(czerr.ErrTransportFailure, "start", err)
	}
	s.cmd = cmd
	s.stdin = stdin
	s.stdout = bufio.NewReader(stdout)
	s.done = make(chan error, 1)
	go func() {
		s.done <- cmd.Wait()
	}()
	return nil
}

func (s *Stdio) Send(ctx context.Context, frame Frame) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stdin == nil {
		return czerr.New(czerr.ErrTransportFailure, "stdio not open")
	}
	if err := ctx.Err(); err != nil {
		return mapCtx(err)
	}
	return protocol.WriteNDJSON(s.stdin, frame)
}

func (s *Stdio) Receive(ctx context.Context) (Frame, error) {
	s.mu.Lock()
	r := s.stdout
	s.mu.Unlock()
	if r == nil {
		return Frame{}, czerr.New(czerr.ErrTransportFailure, "stdio not open")
	}
	type result struct {
		f   Frame
		err error
	}
	ch := make(chan result, 1)
	go func() {
		f, err := protocol.ReadNDJSON(r)
		ch <- result{f, err}
	}()
	select {
	case <-ctx.Done():
		return Frame{}, mapCtx(ctx.Err())
	case err := <-s.done:
		detail := strings.TrimSpace(s.stderr.String())
		if detail != "" {
			if err != nil {
				return Frame{}, czerr.Wrap(czerr.ErrBridgeFailed, "process exited: "+detail, err)
			}
			return Frame{}, czerr.New(czerr.ErrBridgeFailed, "process exited: "+detail)
		}
		if err != nil {
			return Frame{}, czerr.Wrap(czerr.ErrBridgeFailed, "process exited", err)
		}
		return Frame{}, czerr.New(czerr.ErrBridgeFailed, "process exited")
	case res := <-ch:
		if res.err != nil {
			if detail := strings.TrimSpace(s.stderr.String()); detail != "" {
				return Frame{}, czerr.Wrap(czerr.ErrTransportFailure, "stdio: "+detail, res.err)
			}
		}
		return res.f, res.err
	}
}

func (s *Stdio) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stdin != nil {
		_ = s.stdin.Close()
	}
	if s.cmd != nil && s.cmd.Process != nil {
		_ = s.cmd.Process.Kill()
	}
	return nil
}

// PID returns the child process id, or 0.
func (s *Stdio) PID() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.cmd == nil || s.cmd.Process == nil {
		return 0
	}
	return s.cmd.Process.Pid
}

// Wait waits for process exit.
func (s *Stdio) Wait() error {
	if s.done == nil {
		return nil
	}
	return <-s.done
}
