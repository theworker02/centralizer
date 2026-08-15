package health

import (
	"fmt"
	"strings"
)

func formatSnapshot(s Snapshot) string {
	var b strings.Builder
	fmt.Fprintf(&b, "Service: %s\n", s.Service)
	fmt.Fprintf(&b, "Health:\n%s\n", strings.ToUpper(string(s.State)))
	fmt.Fprintf(&b, "Transport:\n%s\n", s.Transport)
	fmt.Fprintf(&b, "Runtime:\n%s\n", s.Runtime)
	if s.Latency > 0 {
		fmt.Fprintf(&b, "Latency:\n%v\n", s.Latency)
	}
	fmt.Fprintf(&b, "Success Rate:\n%.2f%%\n", s.SuccessRate*100)
	fmt.Fprintf(&b, "Restarts:\n%d\n", s.Restarts)
	fmt.Fprintf(&b, "Fallback Count:\n%d\n", s.Fallbacks)
	return b.String()
}
