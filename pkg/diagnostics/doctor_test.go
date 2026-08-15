package diagnostics

import "testing"

func TestRunIncludesCoreChecks(t *testing.T) {
	rep := Run([]string{"python", "go"})
	if rep.Version == "" || rep.OS == "" {
		t.Fatalf("%+v", rep)
	}
	want := map[string]bool{"cache": false, "Git": false, "protocol": false, "adapters": false}
	for _, c := range rep.Checks {
		if _, ok := want[c.Name]; ok {
			want[c.Name] = true
		}
		if c.Name == "protocol" && !c.OK {
			t.Fatalf("protocol: %+v", c)
		}
		if c.Name == "cache" && !c.OK {
			t.Fatalf("cache: %+v", c)
		}
	}
	for name, seen := range want {
		if !seen {
			t.Fatalf("missing check %s", name)
		}
	}
	if rep.Text() == "" {
		t.Fatal("empty text")
	}
}
