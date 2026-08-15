package adapter

import (
	"context"
	"testing"

	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/capability"
	"github.com/theworker02/centralizer/pkg/czerr"
)

type stubAdapter struct{}

func (stubAdapter) Name() string { return "python" }
func (stubAdapter) Detect(context.Context, Target) (Detection, error) {
	return Detection{}, czerr.ErrUnsupportedTarget
}
func (stubAdapter) Capabilities(context.Context, Target) ([]capability.Capability, error) {
	return []capability.Capability{{Kind: capability.Stdio, Available: true}}, nil
}
func (stubAdapter) Prepare(context.Context, Target) error { return nil }
func (stubAdapter) Connect(context.Context, Target, bridge.Plan) (bridge.Bridge, error) {
	return nil, czerr.ErrNotImplemented
}

func TestCatalog(t *testing.T) {
	r := NewRegistry()
	r.Register(stubAdapter{})
	cat := Catalog(r)
	if len(cat) != 1 {
		t.Fatalf("len=%d", len(cat))
	}
	if cat[0].Name != "python" || cat[0].Tier != 1 || !cat[0].Invocation {
		t.Fatalf("%+v", cat[0])
	}
	if len(cat[0].Capabilities) != 1 || cat[0].Capabilities[0] != "stdio" {
		t.Fatalf("caps=%v", cat[0].Capabilities)
	}
}
