package entities

import "testing"

func TestFlattenedNucleiVarsStableOrder(t *testing.T) {
	request := ScanRequest{
		RequestParams: map[string]string{
			"b": "two",
			"a": "one",
		},
	}

	got := request.FlattenedNucleiVars()
	if len(got) != 2 {
		t.Fatalf("expected 2 params, got %d", len(got))
	}
	if got[0] != "-var a=one" || got[1] != "-var b=two" {
		t.Fatalf("unexpected flattened params: %#v", got)
	}
}
