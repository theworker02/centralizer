//go:build windows

package transport

import (
	"context"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/theworker02/centralizer/pkg/czerr"
)

var (
	kernel32      = syscall.NewLazyDLL("kernel32.dll")
	createFileW   = kernel32.NewProc("CreateFileW")
	invalidHandle = ^uintptr(0)
)

// Open dials an existing named pipe. Creating a server-side pipe is
// experimental and not implemented in this release.
func (p *NamedPipe) Open(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return mapCtx(err)
	}
	path := p.Name
	if path == "" {
		return czerr.New(czerr.ErrInvalidArgument, "named pipe name required")
	}
	if !strings.HasPrefix(strings.ToLower(path), `\\.\pipe\`) {
		path = `\\.\pipe\` + path
	}
	ptr, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return czerr.Wrap(czerr.ErrTransportFailure, "named_pipe", err)
	}
	handle, _, callErr := createFileW.Call(
		uintptr(unsafe.Pointer(ptr)),
		syscall.GENERIC_READ|syscall.GENERIC_WRITE,
		0,
		0,
		syscall.OPEN_EXISTING,
		syscall.FILE_ATTRIBUTE_NORMAL,
		0,
	)
	if handle == 0 || handle == invalidHandle {
		return czerr.Wrap(czerr.ErrTransportFailure, "named_pipe dial "+path, callErr)
	}
	p.mu.Lock()
	p.conn = os.NewFile(handle, path)
	p.mu.Unlock()
	return nil
}
