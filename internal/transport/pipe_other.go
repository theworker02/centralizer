//go:build !windows

package transport

import (
	"context"

	"github.com/theworker02/centralizer/pkg/czerr"
)

// Open reports that named pipes are a Windows transport.
func (p *NamedPipe) Open(context.Context) error {
	return czerr.New(czerr.ErrNotImplemented, "named pipes are a Windows transport")
}
