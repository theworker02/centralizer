package bridge

import "testing"

func TestTransportName(t *testing.T) {
	cases := map[Strategy]string{
		StrategyInProcess:    "native",
		StrategyUnixSocket:   "unix_socket",
		StrategyNamedPipe:    "named_pipe",
		StrategyStdio:        "stdio",
		StrategyTCP:          "tcp",
		StrategyWASM:         "wasm",
		StrategySharedMemory: "shared_memory",
	}
	for strategy, want := range cases {
		if got := TransportName(strategy); got != want {
			t.Fatalf("%s: got %q want %q", strategy, got, want)
		}
	}
	if TransportName(Strategy("unknown")) != "unknown" {
		t.Fatal("unknown strategies should pass through")
	}
}
