import { useEffect, useState } from "react";

const pages = [
  "overview",
  "quickstart",
  "architecture",
  "cir",
  "planner",
  "adapters",
  "api",
  "cli",
  "protocol",
  "security",
  "benchmarks",
  "roadmap",
  "changelog",
  "privacy",
] as const;

type Page = (typeof pages)[number];

const labels: Record<Page, string> = {
  overview: "Overview",
  quickstart: "Quickstart",
  architecture: "Architecture",
  cir: "CIR",
  planner: "Planner",
  adapters: "Adapters",
  api: "Go API",
  cli: "CLI",
  protocol: "Protocol",
  security: "Security",
  benchmarks: "Benchmarks",
  roadmap: "Roadmap",
  changelog: "Changelog",
  privacy: "Privacy",
};

export function App() {
  const [page, setPage] = useState<Page>(readHash());

  useEffect(() => {
    const onHash = () => setPage(readHash());
    window.addEventListener("hashchange", onHash);
    return () => window.removeEventListener("hashchange", onHash);
  }, []);

  return (
    <div className="shell">
      <nav>
        <div className="brand">
          <img src="./logo.svg" alt="" />
          <div>
            <strong>CENTRALIZER</strong>
            <div className="brand-sub">interop runtime</div>
          </div>
        </div>
        {pages.map((id) => (
          <a
            key={id}
            className={page === id ? "item active" : "item"}
            href={`#${id}`}
          >
            {labels[id]}
          </a>
        ))}
        <div className="nav-foot">
          <a href="https://github.com/theworker02/centralizer">GitHub</a>
          <a href="https://pkg.go.dev/github.com/theworker02/centralizer/pkg/centralizer">
            Go API
          </a>
          <a href="#privacy">Privacy</a>
        </div>
      </nav>
      <main>{render(page)}</main>
    </div>
  );
}

function readHash(): Page {
  const h = window.location.hash.replace("#", "") as Page;
  return pages.includes(h) ? h : "overview";
}

function render(page: Page) {
  switch (page) {
    case "overview":
      return <Overview />;
    case "quickstart":
      return <Quickstart />;
    case "architecture":
      return <Architecture />;
    case "cir":
      return <CIR />;
    case "planner":
      return <Planner />;
    case "adapters":
      return <Adapters />;
    case "api":
      return <API />;
    case "cli":
      return <CLI />;
    case "protocol":
      return <Protocol />;
    case "security":
      return <Security />;
    case "benchmarks":
      return <Benchmarks />;
    case "roadmap":
      return <Roadmap />;
    case "changelog":
      return <Changelog />;
    case "privacy":
      return <Privacy />;
  }
}

function Overview() {
  return (
    <>
      <img className="hero-mark" src="./logo.svg" alt="Centralizer" />
      <div className="kicker">Interop runtime · v0.1.2</div>
      <h1>CENTRALIZER</h1>
      <p className="tag">One runtime. Every language.</p>
      <p>
        Centralizer is a Go-based interoperability runtime that discovers,
        connects, supervises, and exposes software written across different
        programming languages through a unified interface. The calling
        application does not need to know which bridge was selected.
      </p>
      <pre>{`hub := centralizer.New()
service, err := hub.Connect(ctx, "./analytics")
result, err := service.Call(ctx, "calculate", centralizer.Args{
    "value": 42,
})`}</pre>
      <p>
        Automation stays deterministic, explainable, and bounded. Every
        selection can be printed with <code>centralizer explain</code>.
      </p>
      <h2>Compatibility</h2>
      <p className="note">
        Only mark a capability complete when tests demonstrate it. Detect-only
        adapters are not Call=Yes.
      </p>
      <table>
        <thead>
          <tr>
            <th>Runtime</th>
            <th>Detect</th>
            <th>Call</th>
            <th>Stream</th>
            <th>Handles</th>
            <th>Native</th>
            <th>Process</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>Go</td>
            <td>Yes</td>
            <td>Yes</td>
            <td>Planned</td>
            <td>Yes</td>
            <td>Yes</td>
            <td>Yes</td>
          </tr>
          <tr>
            <td>Python</td>
            <td>Yes</td>
            <td>Yes</td>
            <td>Yes</td>
            <td>Yes</td>
            <td>N/A</td>
            <td>Yes</td>
          </tr>
          <tr>
            <td>Node.js</td>
            <td>Yes</td>
            <td>Yes</td>
            <td>Yes</td>
            <td>Planned</td>
            <td>N/A</td>
            <td>Yes</td>
          </tr>
          <tr>
            <td>Rust</td>
            <td>Yes</td>
            <td>Yes*</td>
            <td>Planned</td>
            <td>Planned</td>
            <td>Planned</td>
            <td>Yes</td>
          </tr>
          <tr>
            <td>C / C++</td>
            <td>Yes</td>
            <td>No</td>
            <td>No</td>
            <td>No</td>
            <td>Planned</td>
            <td>Planned</td>
          </tr>
          <tr>
            <td>WASM</td>
            <td>Yes</td>
            <td>No</td>
            <td>No</td>
            <td>No</td>
            <td>N/A</td>
            <td>N/A</td>
          </tr>
          <tr>
            <td>JVM / .NET / Ruby / PHP</td>
            <td>Yes</td>
            <td>No</td>
            <td>No</td>
            <td>No</td>
            <td>No</td>
            <td>Planned</td>
          </tr>
          <tr>
            <td>Swift / Dart / Lua / Zig</td>
            <td>Yes</td>
            <td>No</td>
            <td>No</td>
            <td>No</td>
            <td>No</td>
            <td>Planned</td>
          </tr>
        </tbody>
      </table>
      <p className="note">
        * Rust invocation requires a target that speaks Centralizer Protocol 1.x
        on stdio.
      </p>
    </>
  );
}

function Quickstart() {
  return (
    <>
      <h2>Quickstart</h2>
      <p>Requires Go 1.23+.</p>
      <pre>{`go get github.com/theworker02/centralizer/pkg/centralizer@latest
go install github.com/theworker02/centralizer/cmd/centralizer@latest`}</pre>
      <p>Connect the Python example:</p>
      <pre>{`centralizer detect ./examples/go-python/analytics
centralizer explain ./examples/go-python/analytics
centralizer call ./examples/go-python/analytics calculate value=21
centralizer lock ./examples/go-python/analytics`}</pre>
      <p>
        <code>centralizer init</code> writes a starter <code>centralizer.yaml</code>.
        An optional <code>schema.yaml</code> next to the target is loaded and
        used to validate calls. Inferred schemas still skip type checks.
      </p>
      <h3>Library</h3>
      <pre>{`package main

import (
    "context"
    "fmt"
    "github.com/theworker02/centralizer/pkg/centralizer"
)

func main() {
    ctx := context.Background()
    hub := centralizer.New()
    analytics, err := hub.Connect(ctx, "./examples/go-python/analytics")
    if err != nil { panic(err) }
    defer analytics.Close(ctx)
    result, err := analytics.Call(ctx, "calculate", centralizer.Args{"value": 42})
    if err != nil { panic(err) }
    fmt.Println(result)
}`}</pre>
    </>
  );
}

function Architecture() {
  return (
    <>
      <h2>Architecture</h2>
      <p>
        Centralizer is a single Go module. Each package has one job. The Hub
        owns services. A Service owns its supervisor. The supervisor owns the
        live Bridge. A stdio or TCP transport owns the child process and must
        kill it on Close.
      </p>
      <div className="flow" aria-label="Architecture">
        <span>Application</span>
        <i />
        <span>Hub API</span>
        <i />
        <span>Discovery</span>
        <i />
        <span>Planner</span>
        <i />
        <span>CIR</span>
        <i />
        <span>Supervisor</span>
        <i />
        <span>Adapter</span>
        <i />
        <span>Target</span>
      </div>
      <table>
        <thead>
          <tr>
            <th>Package</th>
            <th>Responsibility</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>
              <code>pkg/centralizer</code>
            </td>
            <td>Public Hub / Service API</td>
          </tr>
          <tr>
            <td>
              <code>pkg/cir</code>
            </td>
            <td>Kind-tagged intermediate values</td>
          </tr>
          <tr>
            <td>
              <code>pkg/schema</code>
            </td>
            <td>Callable surface; explicit YAML load</td>
          </tr>
          <tr>
            <td>
              <code>internal/planner</code>
            </td>
            <td>Deterministic strategy selection</td>
          </tr>
          <tr>
            <td>
              <code>internal/supervisor</code>
            </td>
            <td>Health, recovery, circuit breaker</td>
          </tr>
          <tr>
            <td>
              <code>internal/transport</code>
            </td>
            <td>stdio, TCP loopback, Unix, named pipe foundation</td>
          </tr>
        </tbody>
      </table>
    </>
  );
}

function CIR() {
  return (
    <>
      <h2>CIR</h2>
      <p>
        Centralizer Intermediate Representation is the semantic boundary
        between languages. Values are kind-tagged. They are not stored as an
        untyped bag of <code>any</code>.
      </p>
      <table>
        <thead>
          <tr>
            <th>Kind</th>
            <th>Notes</th>
          </tr>
        </thead>
        <tbody>
          <tr>
            <td>null, boolean, int, uint, float, decimal</td>
            <td>scalars; non-finite floats rejected by schema validate</td>
          </tr>
          <tr>
            <td>string, bytes</td>
            <td>text and raw</td>
          </tr>
          <tr>
            <td>array, tuple, map, struct</td>
            <td>composites</td>
          </tr>
          <tr>
            <td>optional, result, error, enum, union</td>
            <td>sum types</td>
          </tr>
          <tr>
            <td>timestamp, duration, UUID</td>
            <td>time and identity</td>
          </tr>
          <tr>
            <td>stream, handle, opaque</td>
            <td>handles never store a foreign pointer</td>
          </tr>
        </tbody>
      </table>
      <pre>{`{"k":"int","i":42}
{"k":"map","m":[{"k":"value","v":{"k":"int","i":42}}]}`}</pre>
    </>
  );
}

function Planner() {
  return (
    <>
      <h2>Bridge planner</h2>
      <p>
        Strategies are scored on performance, isolation, compatibility, and
        policy. Equivalent inputs produce the same ranking. Policy always wins
        over a higher benchmark score.
      </p>
      <ol>
        <li>Discover markers (go.mod, pyproject.toml, package.json, Cargo.toml, …).</li>
        <li>Score language/runtime hypotheses. Ambiguity is reported, not hidden.</li>
        <li>Build a capability graph (stdio, process, native, TCP, …).</li>
        <li>Score strategies. Select deterministically. Record fallbacks.</li>
        <li>
          Explain the decision with <code>centralizer explain</code>.
        </li>
      </ol>
      <p>
        Optional <code>centralizer.lock</code> snapshots the resolved adapter,
        transport, fingerprint, and scores. Connect does not require it.
      </p>
    </>
  );
}

function Adapters() {
  return (
    <>
      <h2>Adapters</h2>
      <p>
        Implement <code>adapter.Adapter</code> and register it on a Hub with{" "}
        <code>WithAdapter</code>. Do not use Go plugins as the extension
        mechanism. Claim only what tests demonstrate.
      </p>
      <pre>{`type Adapter interface {
    Name() string
    Detect(context.Context, Target) (Detection, error)
    Capabilities(context.Context, Target) ([]capability.Capability, error)
    Prepare(context.Context, Target) error
    Connect(context.Context, Target, bridge.Plan) (bridge.Bridge, error)
}`}</pre>
      <p>
        <code>centralizer adapters</code> lists tier and claimed capabilities.
        Foundation adapters (C, WASM, JVM, …) detect only; Connect returns not
        implemented.
      </p>
    </>
  );
}

function API() {
  return (
    <>
      <h2>Go API</h2>
      <pre>{`hub := centralizer.New(
    centralizer.WithAutoRecovery(true),
    centralizer.WithTracing(true),
    centralizer.WithHandleTTL(time.Hour),
)
svc, err := hub.Connect(ctx, "./analytics")
_ = svc.Language()
_ = svc.Health()
_ = svc.Plan()
st, err := svc.Stream(ctx, "count_up", centralizer.Args{"n": 4})
h, err := svc.New(ctx, "Session", nil)
_ = svc.Release(ctx, id)`}</pre>
      <p>
        Public APIs do not panic. Close drops locally tracked handles via
        DropBridge and reaps child processes.
      </p>
    </>
  );
}

function CLI() {
  return (
    <>
      <h2>CLI</h2>
      <p className="mono">
        detect · inspect · describe · connect · call · health · list · graph ·
        explain · bench · trace · doctor · cache · init · lock · adapters ·
        version
      </p>
      <p>Global flags: --json --quiet --verbose</p>
      <pre>{`centralizer init
centralizer adapters --json
centralizer lock ./examples/go-python/analytics
centralizer doctor`}</pre>
      <p>
        <code>centralizerd</code> is optional.{" "}
        <code>GET /v1/metrics</code> returns JSON counters on loopback.
      </p>
    </>
  );
}

function Protocol() {
  return (
    <>
      <h2>Protocol 1.0</h2>
      <p>
        NDJSON on stdio. Length-prefixed JSON on TCP and Unix sockets. HELLO
        negotiates the major version. Maximum frame size is 16 MiB.
      </p>
      <pre>{`{"v":1,"id":"2","type":"CALL","payload":{"function":"calculate","args":{"value":{"k":"int","i":42}}}}
{"v":1,"id":"3","type":"STREAM_OPEN","payload":{"name":"count_up","stream":"st-1"}}
{"v":1,"id":"st-1","type":"STREAM_DATA","payload":{"stream":"st-1","value":{"k":"int","i":0}}}`}</pre>
      <p>
        Python and Node generated shims speak STREAM_OPEN / STREAM_DATA /
        STREAM_CLOSE for generator and iterable functions. Prefer{" "}
        <code>tcp</code> in the manifest to use the localhost framed transport.
      </p>
    </>
  );
}

function Security() {
  return (
    <>
      <h2>Security</h2>
      <p>
        Discovered code is not trusted because it was found. Paths are
        validated, environments filtered, payloads size-limited, and the daemon
        binds loopback only. Handles never store foreign memory pointers.
        Report vulnerabilities privately via GitHub Security Advisories.
      </p>
      <p>
        Centralizer does not require accounts and does not ship analytics to
        the authors. See <a href="#privacy">Privacy</a> and{" "}
        <a href="https://github.com/theworker02/centralizer/blob/main/PRIVACY.md">
          PRIVACY.md
        </a>
        .
      </p>
    </>
  );
}

function Benchmarks() {
  return (
    <>
      <h2>Benchmarks</h2>
      <p>
        Publish numbers only with machine, CPU, OS, runtime, Centralizer
        version, test, sample size, transport, and payload size.{" "}
        <code>centralizer bench</code> reports planner scores and does not
        override policy.
      </p>
    </>
  );
}

function Roadmap() {
  return (
    <>
      <h2>Roadmap</h2>
      <p>
        v0.1 is a vertical slice: Go host, Python / Node / Rust process
        targets, CIR, planner, supervisor, CLI. Streaming shims, lock files,
        explicit schemas, localhost TCP, and handle expiry are in this tree.
        Later phases add C/WASM invocation, JVM/.NET, shared memory, and a
        multi-process daemon registry.
      </p>
      <p className="note">
        Do not treat later phases as implemented because files exist.
      </p>
    </>
  );
}

function Privacy() {
  return (
    <>
      <h2>Privacy</h2>
      <p>
        Centralizer is a local library and CLI. It does not require an
        account. There is no Centralizer cloud and no sign-in.
      </p>
      <ul>
        <li>
          No analytics or telemetry is shipped to the authors by default.
        </li>
        <li>
          Optional <code>centralizerd</code> binds loopback only.
        </li>
        <li>Doctor reports and the cache stay on this machine.</li>
        <li>
          Connecting a target runs that code locally. You are responsible for
          it.
        </li>
        <li>
          GitHub, pkg.go.dev, and proxy.golang.org are third parties if you
          use those services.
        </li>
      </ul>
      <p>
        Full text:{" "}
        <a href="https://github.com/theworker02/centralizer/blob/main/PRIVACY.md">
          PRIVACY.md
        </a>
        . Security reports go through{" "}
        <a href="https://github.com/theworker02/centralizer/security/advisories">
          GitHub Security Advisories
        </a>
        , not public issues.
      </p>
    </>
  );
}

function Changelog() {
  return (
    <>
      <h2>Changelog</h2>
      <h3>0.1.2</h3>
      <ul>
        <li>Privacy policy (PRIVACY.md) and site privacy page</li>
        <li>Expanded README: module identity, errors, cache, versioning</li>
        <li>Tagged module for pkg.go.dev / proxy.golang.org</li>
      </ul>
      <h3>0.1.1</h3>
      <ul>
        <li>Official brand package and assets/ as the single logo source</li>
        <li>Supervisor fallback uses TransportName (in_process → native)</li>
        <li>Expanded README: CLI, CIR, planner, protocol, security</li>
      </ul>
      <h3>0.1.0</h3>
      <p>
        Go hub API, CIR, Protocol 1.0, discovery, planner, supervisor, Python /
        Node / Rust adapters, CLI, optional daemon, tests, Apache 2.0.
      </p>
    </>
  );
}
