package shim

import (
	"context"
	"os"
	"os/exec"

	"github.com/theworker02/centralizer/internal/security"
	"github.com/theworker02/centralizer/internal/session"
	"github.com/theworker02/centralizer/internal/transport"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/czerr"
)

// StdioConfig launches a protocol-speaking child process.
type StdioConfig struct {
	Argv []string
	Dir  string
	Env  []string
	Plan bridge.Plan
}

// ConnectStdio starts the process and completes the protocol handshake.
func ConnectStdio(ctx context.Context, cfg StdioConfig) (bridge.Bridge, *transport.Stdio, error) {
	env := cfg.Env
	if env == nil {
		env = os.Environ()
	}
	tr := &transport.Stdio{
		Name: cfg.Argv[0],
		Argv: cfg.Argv,
		Dir:  cfg.Dir,
		Env:  env,
	}
	sess, err := session.Open(ctx, tr, cfg.Plan)
	if err != nil {
		_ = tr.Close()
		return nil, nil, err
	}
	return sess, tr, nil
}

// Connect selects stdio or localhost TCP from the plan transport.
func Connect(ctx context.Context, cfg StdioConfig) (bridge.Bridge, error) {
	if cfg.Plan.Transport == "tcp" {
		return ConnectTCP(ctx, cfg)
	}
	b, _, err := ConnectStdio(ctx, cfg)
	return b, err
}

// ConnectTCP listens on 127.0.0.1, starts the child, and handshakes over
// length-prefixed frames. The child must honor CENTRALIZER_TRANSPORT=tcp
// and CENTRALIZER_ADDR.
func ConnectTCP(ctx context.Context, cfg StdioConfig) (bridge.Bridge, error) {
	if len(cfg.Argv) == 0 {
		return nil, czerr.New(czerr.ErrTransportFailure, "empty argv")
	}
	ln, err := transport.ListenLoopback()
	if err != nil {
		return nil, err
	}
	addr := ln.Addr().String()

	env := append([]string{}, cfg.Env...)
	if env == nil {
		env = os.Environ()
	}
	env = append(env, "CENTRALIZER_TRANSPORT=tcp", "CENTRALIZER_ADDR="+addr)

	cmd := exec.CommandContext(ctx, cfg.Argv[0], cfg.Argv[1:]...)
	cmd.Dir = cfg.Dir
	cmd.Env = security.FilterEnv(env)
	if err := cmd.Start(); err != nil {
		_ = ln.Close()
		return nil, czerr.Wrap(czerr.ErrTransportFailure, "start tcp child", err)
	}

	conn, err := transport.AcceptOne(ctx, ln)
	_ = ln.Close()
	if err != nil {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
		return nil, err
	}

	tr := &ownedNet{Net: transport.NewConn(conn, transport.KindTCP), cmd: cmd}
	sess, err := session.Open(ctx, tr, cfg.Plan)
	if err != nil {
		_ = tr.Close()
		return nil, err
	}
	return sess, nil
}

type ownedNet struct {
	*transport.Net
	cmd *exec.Cmd
}

func (o *ownedNet) Close() error {
	err := o.Net.Close()
	if o.cmd != nil && o.cmd.Process != nil {
		_ = o.cmd.Process.Kill()
	}
	return err
}
