package centralizer

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/theworker02/centralizer/internal/session"
	"github.com/theworker02/centralizer/pkg/cir"
	"github.com/theworker02/centralizer/pkg/czerr"
)

func TestNativeCall(t *testing.T) {
	hub := New(WithAutoRecovery(false), WithTimeout(2*time.Second))
	hub.RegisterNative(&session.Handler{
		Name: "math",
		Funcs: map[string]session.Func{
			"calculate": func(_ context.Context, args map[string]cir.Value) (cir.Value, error) {
				v, ok := args["value"]
				if !ok {
					return cir.Value{}, czerr.New(czerr.ErrInvalidArgument, "value")
				}
				n, err := v.Int()
				if err != nil {
					return cir.Value{}, err
				}
				return cir.Int(n * 2), nil
			},
		},
	})
	ctx := context.Background()
	svc, err := hub.Connect(ctx, "native:math")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = svc.Close(ctx) }()
	if svc.Language() != "Go" {
		t.Fatalf("language=%s", svc.Language())
	}
	got, err := svc.Call(ctx, "calculate", Args{"value": 21})
	if err != nil {
		t.Fatal(err)
	}
	n, err := got.Int()
	if err != nil || n != 42 {
		t.Fatalf("got %v %v", n, err)
	}
	h := svc.Health()
	if h.Calls < 1 {
		t.Fatalf("health=%+v", h)
	}
}

func TestDetectPythonFixture(t *testing.T) {
	root := filepath.Join("..", "..", "examples", "go-python", "analytics")
	if _, err := os.Stat(root); err != nil {
		t.Skip("example missing")
	}
	hub := New()
	res, err := hub.Analyze(context.Background(), root)
	if err != nil {
		t.Fatal(err)
	}
	if res.Primary.Adapter != "python" {
		t.Fatalf("primary=%+v", res.Primary)
	}
	if res.Primary.Confidence < 0.5 {
		t.Fatalf("confidence=%f", res.Primary.Confidence)
	}
}

func TestLockPlanNative(t *testing.T) {
	hub := New(WithAutoRecovery(false))
	hub.RegisterNative(&session.Handler{Name: "math", Funcs: map[string]session.Func{
		"n": func(context.Context, map[string]cir.Value) (cir.Value, error) { return cir.Int(1), nil },
	}})
	lf, err := hub.LockPlan(context.Background(), "native:math")
	if err != nil {
		t.Fatal(err)
	}
	if lf.Adapter != "go" || lf.Transport != "native" {
		t.Fatalf("%+v", lf)
	}
	cat := hub.AdapterCatalog()
	if len(cat) < 4 {
		t.Fatalf("catalog=%d", len(cat))
	}
}

func TestSplitRefDriveLetterAndEntry(t *testing.T) {
	name, path, entry := splitRef(`C:\Users\proj`)
	if entry != "" {
		t.Fatalf("drive letter must not become an entry: name=%q path=%q entry=%q", name, path, entry)
	}
	if path != `C:\Users\proj` {
		t.Fatalf("path=%q", path)
	}
	_, path, entry = splitRef(`C:\Users\proj:calculate`)
	if path != `C:\Users\proj` || entry != "calculate" {
		t.Fatalf("entry split: path=%q entry=%q", path, entry)
	}
	_, path, entry = splitRef("./analytics:calculate")
	if path != "./analytics" || entry != "calculate" {
		t.Fatalf("unix-style: path=%q entry=%q", path, entry)
	}
	name, path, entry = splitRef("native:math")
	if name != "math" || path != "native:math" || entry != "math" {
		t.Fatalf("native: name=%q path=%q entry=%q", name, path, entry)
	}
}

func TestHandleExpiryAndDrop(t *testing.T) {
	hub := New(WithAutoRecovery(false), WithHandleTTL(8*time.Millisecond))
	hub.RegisterNative(&session.Handler{
		Name: "box",
		Types: map[string]session.Constructor{
			"Box": func(context.Context, map[string]cir.Value) (any, error) {
				return struct{}{}, nil
			},
		},
	})
	ctx := context.Background()
	svc, err := hub.Connect(ctx, "native:box")
	if err != nil {
		t.Fatal(err)
	}
	h, err := svc.New(ctx, "Box", nil)
	if err != nil {
		t.Fatal(err)
	}
	id, err := h.HandleID()
	if err != nil {
		t.Fatal(err)
	}
	time.Sleep(20 * time.Millisecond)
	_, err = svc.Get(ctx, id, "x")
	if !errors.Is(err, czerr.ErrHandleInvalid) {
		t.Fatalf("expected expired handle, got %v", err)
	}
	if err := svc.Close(ctx); err != nil {
		t.Fatal(err)
	}
}
