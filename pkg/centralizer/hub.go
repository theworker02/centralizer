package centralizer

import (
	"context"
	"path/filepath"
	"strings"
	"sync"

	found "github.com/theworker02/centralizer/adapters/foundation"
	adpgo "github.com/theworker02/centralizer/adapters/golang"
	adpnode "github.com/theworker02/centralizer/adapters/node"
	adppy "github.com/theworker02/centralizer/adapters/python"
	adprust "github.com/theworker02/centralizer/adapters/rust"
	"github.com/theworker02/centralizer/internal/cache"
	"github.com/theworker02/centralizer/internal/discovery"
	"github.com/theworker02/centralizer/internal/lifecycle"
	"github.com/theworker02/centralizer/internal/planner"
	"github.com/theworker02/centralizer/internal/registry"
	"github.com/theworker02/centralizer/internal/security"
	"github.com/theworker02/centralizer/internal/session"
	"github.com/theworker02/centralizer/internal/supervisor"
	"github.com/theworker02/centralizer/internal/telemetry"
	"github.com/theworker02/centralizer/pkg/adapter"
	"github.com/theworker02/centralizer/pkg/bridge"
	"github.com/theworker02/centralizer/pkg/czerr"
	"github.com/theworker02/centralizer/pkg/health"
	"github.com/theworker02/centralizer/pkg/lockfile"
	"github.com/theworker02/centralizer/pkg/manifest"
	"github.com/theworker02/centralizer/pkg/schema"
)

// Hub is the process-local Centralizer runtime.
type Hub struct {
	cfg      config
	reg      *adapter.Registry
	services *registry.Registry
	cache    *cache.Store
	goAdp    *adpgo.Adapter
	shutdown lifecycle.Group
	mu       sync.Mutex
	closed   bool
	conns    []*Service
}

// New constructs a Hub with built-in adapters.
func New(opts ...Option) *Hub {
	cfg := defaultConfig()
	for _, o := range opts {
		o(&cfg)
	}
	if cfg.logger != nil {
		telemetry.SetLogger(cfg.logger)
	}
	store, err := cache.New(cfg.cacheDir)
	if err != nil {
		store, _ = cache.New("")
	}
	h := &Hub{
		cfg:      cfg,
		reg:      adapter.NewRegistry(),
		services: registry.New(),
		cache:    store,
		goAdp:    &adpgo.Adapter{},
	}
	h.reg.Register(&adppy.Adapter{Store: store})
	h.reg.Register(&adpnode.Adapter{Store: store})
	h.reg.Register(h.goAdp)
	h.reg.Register(&adprust.Adapter{})
	for _, a := range found.All() {
		h.reg.Register(a)
	}
	for _, a := range cfg.adapters {
		h.reg.Register(a)
	}
	return h
}

// RegisterNative installs an in-process Go handler.
func (h *Hub) RegisterNative(handler *session.Handler) {
	h.goAdp.Register(handler)
}

// Connect discovers target, plans a bridge, and returns a Service.
func (h *Hub) Connect(ctx context.Context, ref string, opts ...Option) (*Service, error) {
	cfg := h.cfg
	for _, o := range opts {
		o(&cfg)
	}
	log := telemetry.FromContext(ctx)
	var tracer *telemetry.Tracer
	if cfg.tracing {
		tracer = telemetry.NewTracer("connect")
		defer tracer.Finish()
	}

	name, path, entry := splitRef(ref)
	if cfg.entry != "" {
		entry = cfg.entry
	}
	schemaRef := ""
	if cfg.manifest != nil {
		if svc, ok := cfg.manifest.Services[name]; ok {
			if path == ref || path == name {
				path = svc.Source
			}
			if entry == "" {
				entry = svc.Entry
			}
			if cfg.language == "" {
				cfg.language = svc.Language
			}
			if len(cfg.prefer) == 0 {
				cfg.prefer = svc.Prefer
			}
			schemaRef = svc.Schema
		}
	}

	resolved := path
	if !strings.HasPrefix(path, "native:") && !strings.HasPrefix(ref, "native:") {
		var err error
		if tracer != nil {
			tracer.Start("discovery")
		}
		resolved, err = security.ResolveTarget(path)
		if tracer != nil {
			tracer.End(err)
		}
		if err != nil {
			return nil, err
		}
	}

	target := adapter.Target{Ref: ref, Path: resolved, Entry: entry, Language: cfg.language}
	eng := &discovery.Engine{Registry: h.reg}
	if tracer != nil {
		tracer.Start("analyze")
	}
	analysis, err := eng.Analyze(ctx, target)
	if tracer != nil {
		tracer.End(err)
	}
	if err != nil {
		return nil, err
	}
	if analysis.Primary.Adapter == "" {
		return nil, czerr.New(czerr.ErrUnsupportedTarget, ref)
	}
	if cfg.language != "" && !strings.EqualFold(cfg.language, "auto") {
		for _, d := range analysis.Detections {
			if strings.EqualFold(d.Language, cfg.language) || strings.EqualFold(d.Adapter, cfg.language) {
				analysis.Primary = d
				break
			}
		}
	}
	pol := security.Engine{Policy: cfg.policy}
	if err := pol.AllowRuntime(analysis.Primary.Adapter); err != nil {
		return nil, err
	}

	adp, ok := h.reg.Get(analysis.Primary.Adapter)
	if !ok {
		return nil, czerr.New(czerr.ErrAdapterFailure, "missing adapter "+analysis.Primary.Adapter)
	}
	if tracer != nil {
		tracer.Start("plan")
	}
	planRes, err := planner.Plan(planner.Input{
		Language: analysis.Primary.Language,
		Runtime:  analysis.Primary.Runtime,
		Adapter:  analysis.Primary.Adapter,
		Graph:    analysis.Graph,
		Policy:   pol,
		Prefer:   cfg.prefer,
	})
	if tracer != nil {
		tracer.End(err)
	}
	if err != nil {
		return nil, err
	}
	telemetry.DefaultMetrics.PlannerDecisions.Add(1)
	plan := planRes.Selected
	plan.Fingerprint = analysis.Fingerprint
	log.Info("bridge selected", "target", ref, "adapter", plan.Adapter, "strategy", plan.Strategy, "score", plan.Scores.Overall)

	if tracer != nil {
		tracer.Start("prepare")
	}
	err = adp.Prepare(ctx, target)
	if tracer != nil {
		tracer.End(err)
	}
	if err != nil {
		return nil, czerr.Wrap(czerr.ErrAdapterFailure, "prepare", err)
	}

	factory := func(ctx context.Context, p bridge.Plan) (bridge.Bridge, error) {
		return adp.Connect(ctx, target, p)
	}
	if tracer != nil {
		tracer.Start("connect")
	}
	inner, err := factory(ctx, plan)
	if tracer != nil {
		tracer.End(err)
	}
	if err != nil {
		return nil, czerr.Wrap(czerr.ErrBridgeUnavailable, plan.Adapter, err)
	}
	if sc, serr := schema.Discover(resolved, schemaRef); serr != nil {
		_ = inner.Close(ctx)
		return nil, serr
	} else if sc != nil {
		if setter, ok := inner.(interface{ SetSchema(*schema.Schema) }); ok {
			setter.SetSchema(sc)
		}
	}
	telemetry.DefaultMetrics.RuntimeStartups.Add(1)

	svcName := name
	if svcName == "" || svcName == ref {
		svcName = filepath.Base(resolved)
	}
	sup := supervisor.New(svcName, inner, plan, factory, supervisor.Config{
		AutoRecover: cfg.autoRecover,
		MaxRestarts: pol.MaxRestarts(),
	})
	svc := &Service{
		hub:       h,
		name:      svcName,
		target:    target,
		analysis:  analysis,
		planRes:   planRes,
		sup:       sup,
		adp:       adp,
		timeout:   cfg.timeout,
		tracer:    tracer,
		handles:   lifecycle.NewTable(),
		handleTTL: cfg.handleTTL,
		bridgeID:  inner.ID(),
	}
	h.mu.Lock()
	h.conns = append(h.conns, svc)
	h.mu.Unlock()
	h.services.Put(&registry.Entry{
		Name:      svcName,
		Target:    resolved,
		Language:  plan.Language,
		Runtime:   plan.Runtime,
		Transport: plan.Transport,
		BridgeID:  inner.ID(),
		Plan:      plan,
		Health:    sup.Snapshot(),
	})
	h.shutdown.Add(svc.Close)
	return svc, nil
}

// Analyze runs discovery without connecting.
func (h *Hub) Analyze(ctx context.Context, ref string) (*discovery.Result, error) {
	path := ref
	if !strings.HasPrefix(ref, "native:") {
		var err error
		path, err = security.ResolveTarget(ref)
		if err != nil {
			return nil, err
		}
	}
	return (&discovery.Engine{Registry: h.reg}).Analyze(ctx, adapter.Target{Ref: ref, Path: path})
}

// Explain returns a human-readable planning report without connecting.
func (h *Hub) Explain(ctx context.Context, ref string, opts ...Option) (string, *planner.Result, error) {
	cfg := h.cfg
	for _, o := range opts {
		o(&cfg)
	}
	analysis, err := h.Analyze(ctx, ref)
	if err != nil {
		return "", nil, err
	}
	res, err := planner.Plan(planner.Input{
		Language: analysis.Primary.Language,
		Runtime:  analysis.Primary.Runtime,
		Adapter:  analysis.Primary.Adapter,
		Graph:    analysis.Graph,
		Policy:   security.Engine{Policy: cfg.policy},
		Prefer:   cfg.prefer,
	})
	if err != nil {
		return "", nil, err
	}
	text := planner.Explain(planner.Input{
		Language: analysis.Primary.Language,
		Runtime:  analysis.Primary.Runtime,
		Graph:    analysis.Graph,
	}, res)
	return text, res, nil
}

// List returns connected services.
func (h *Hub) List() []*registry.Entry { return h.services.List() }

// Health returns health snapshots for connected services.
func (h *Hub) Health() []health.Snapshot {
	h.mu.Lock()
	defer h.mu.Unlock()
	out := make([]health.Snapshot, 0, len(h.conns))
	for _, s := range h.conns {
		out = append(out, s.Health())
	}
	return out
}

// Registry exposes the in-process service table.
func (h *Hub) Registry() *registry.Registry { return h.services }

// Adapters returns registered adapter names.
func (h *Hub) Adapters() []string { return h.reg.Names() }

// AdapterCatalog returns tier, invocation, and claimed capabilities.
func (h *Hub) AdapterCatalog() []adapter.Info { return adapter.Catalog(h.reg) }

// LockPlan snapshots the resolved plan without connecting.
func (h *Hub) LockPlan(ctx context.Context, ref string, opts ...Option) (lockfile.File, error) {
	_, plan, err := h.Explain(ctx, ref, opts...)
	if err != nil {
		return lockfile.File{}, err
	}
	analysis, err := h.Analyze(ctx, ref)
	if err != nil {
		return lockfile.File{}, err
	}
	fp := ""
	if analysis != nil {
		fp = analysis.Fingerprint
	}
	lang, runtime := "", ""
	if analysis != nil {
		lang = analysis.Primary.Language
		runtime = analysis.Primary.Runtime
	}
	return lockfile.FromPlan(ref, lang, runtime, fp, plan.Selected), nil
}

// Cache returns the artifact cache.
func (h *Hub) Cache() *cache.Store { return h.cache }

// Close shuts down every connected service.
func (h *Hub) Close(ctx context.Context) error {
	h.mu.Lock()
	h.closed = true
	h.mu.Unlock()
	return h.shutdown.Shutdown(ctx)
}

func splitRef(ref string) (name, path, entry string) {
	if strings.HasPrefix(ref, "native:") {
		name = strings.TrimPrefix(ref, "native:")
		return name, ref, name
	}
	path = ref
	name = filepath.Base(strings.TrimSuffix(ref, string(filepath.Separator)))
	i := strings.LastIndex(ref, ":")
	if i < 0 || i+1 >= len(ref) {
		return name, path, entry
	}
	// A Windows drive prefix (C:) is not an entry separator. The previous
	// `i > 1` guard skipped this case, but left a dead `i == 1` branch.
	if i == 1 && isDriveLetter(ref[0]) {
		return name, path, entry
	}
	path = ref[:i]
	entry = ref[i+1:]
	name = filepath.Base(path)
	return name, path, entry
}

func isDriveLetter(c byte) bool {
	return (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
}

// LoadManifest is a convenience wrapper around manifest.Load.
func LoadManifest(path string) (*manifest.Manifest, error) {
	return manifest.Load(path)
}
