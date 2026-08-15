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
	for strat, want := range cases {
		if got := TransportName(strat); got != want {
			t.Fatalf("%s: got %q want %q", strat, got, want)
		}
	}
	if TransportName(Strategy("unknown")) != "unknown" {
		t.Fatal("unknown strategies should pass through")
	}
}
