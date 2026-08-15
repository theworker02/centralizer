package integration

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/theworker02/centralizer/pkg/centralizer"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		next := filepath.Dir(dir)
		if next == dir {
			t.Fatal("go.mod not found")
		}
		dir = next
	}
}

func TestGoPython(t *testing.T) {
	if _, err := exec.LookPath("python"); err != nil {
		if _, err := exec.LookPath("python3"); err != nil {
			t.Skip("python not installed")
		}
	}
	root := repoRoot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	hub := centralizer.New(centralizer.WithCacheDir(t.TempDir()), centralizer.WithTimeout(15*time.Second))
	svc, err := hub.Connect(ctx, filepath.Join(root, "examples", "go-python", "analytics"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = svc.Close(ctx) }()
	v, err := svc.Call(ctx, "calculate", centralizer.Args{"value": 21})
	if err != nil {
		t.Fatal(err)
	}
	f, err := v.Float()
	if err != nil || f != 42 {
		t.Fatalf("got %v %v", f, err)
	}
}

func TestGoNode(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not installed")
	}
	root := repoRoot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	hub := centralizer.New(centralizer.WithCacheDir(t.TempDir()), centralizer.WithTimeout(15*time.Second))
	svc, err := hub.Connect(ctx, filepath.Join(root, "examples", "go-node", "reporter"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = svc.Close(ctx) }()
	v, err := svc.Call(ctx, "ping", nil)
	if err != nil {
		t.Fatal(err)
	}
	s, err := v.AsString()
	if err != nil || s != "ok" {
		t.Fatalf("got %q %v", s, err)
	}
}

func TestGoPythonStream(t *testing.T) {
	if _, err := exec.LookPath("python"); err != nil {
		if _, err := exec.LookPath("python3"); err != nil {
			t.Skip("python not installed")
		}
	}
	root := repoRoot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	hub := centralizer.New(centralizer.WithCacheDir(t.TempDir()), centralizer.WithTimeout(15*time.Second))
	svc, err := hub.Connect(ctx, filepath.Join(root, "examples", "go-python", "analytics"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = svc.Close(ctx) }()
	st, err := svc.Stream(ctx, "count_up", centralizer.Args{"n": 3})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	var got []int64
	for v := range st.Values() {
		n, err := v.Int()
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, n)
	}
	if len(got) != 3 || got[0] != 0 || got[2] != 2 {
		t.Fatalf("got %v err=%v", got, st.Err())
	}
}

func TestGoPythonTCP(t *testing.T) {
	if _, err := exec.LookPath("python"); err != nil {
		if _, err := exec.LookPath("python3"); err != nil {
			t.Skip("python not installed")
		}
	}
	root := repoRoot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	hub := centralizer.New(
		centralizer.WithCacheDir(t.TempDir()),
		centralizer.WithTimeout(15*time.Second),
		centralizer.WithPrefer("tcp"),
	)
	svc, err := hub.Connect(ctx, filepath.Join(root, "examples", "go-python", "analytics"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = svc.Close(ctx) }()
	if svc.Transport() != "tcp" {
		t.Fatalf("transport=%s", svc.Transport())
	}
	v, err := svc.Call(ctx, "calculate", centralizer.Args{"value": 21})
	if err != nil {
		t.Fatal(err)
	}
	f, err := v.Float()
	if err != nil || f != 42 {
		t.Fatalf("got %v %v", f, err)
	}
}

func TestGoNodeStream(t *testing.T) {
	if _, err := exec.LookPath("node"); err != nil {
		t.Skip("node not installed")
	}
	root := repoRoot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	hub := centralizer.New(centralizer.WithCacheDir(t.TempDir()), centralizer.WithTimeout(15*time.Second))
	svc, err := hub.Connect(ctx, filepath.Join(root, "examples", "go-node", "reporter"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = svc.Close(ctx) }()
	st, err := svc.Stream(ctx, "countUp", centralizer.Args{"n": 3})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()
	var n int
	for range st.Values() {
		n++
	}
	if n != 3 {
		t.Fatalf("got %d values err=%v", n, st.Err())
	}
}

func TestExplicitSchema(t *testing.T) {
	if _, err := exec.LookPath("python"); err != nil {
		if _, err := exec.LookPath("python3"); err != nil {
			t.Skip("python not installed")
		}
	}
	root := repoRoot(t)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	hub := centralizer.New(centralizer.WithCacheDir(t.TempDir()), centralizer.WithTimeout(15*time.Second))
	svc, err := hub.Connect(ctx, filepath.Join(root, "examples", "go-python", "analytics"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = svc.Close(ctx) }()
	sc, err := svc.Describe(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if sc.Inferred {
		t.Fatal("expected explicit schema.yaml to be loaded")
	}
	_, err = svc.Call(ctx, "missing_fn", nil)
	if err == nil {
		t.Fatal("expected schema mismatch")
	}
}

func TestDetectC(t *testing.T) {
	root := repoRoot(t)
	hub := centralizer.New()
	res, err := hub.Analyze(context.Background(), filepath.Join(root, "examples", "go-c"))
	if err != nil {
		t.Fatal(err)
	}
	if res.Primary.Adapter != "c" && res.Primary.Language != "C" {
		t.Fatalf("expected C, got %+v", res.Primary)
	}
}
