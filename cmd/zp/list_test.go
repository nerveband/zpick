package main

import (
	"encoding/json"
	"testing"

	"github.com/nerveband/zpick/internal/backend"
)

// Golden test: verify the JSON contract for list --json.
// These fields must remain present for backwards compatibility (additive only).
func TestListJSONContract(t *testing.T) {
	result := ListResult{
		Sessions: []backend.Session{
			{Name: "test", PID: 123, Clients: 1, StartedIn: "~/test", Active: true},
		},
		Count:          1,
		ZmoshVersion:   "0.4.2",
		BackendVersion: "0.4.2",
	}
	b, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	var raw map[string]interface{}
	if err := json.Unmarshal(b, &raw); err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}

	// Required top-level fields (zmosh_version kept for backwards compat)
	for _, field := range []string{"sessions", "count", "zmosh_version"} {
		if _, ok := raw[field]; !ok {
			t.Errorf("missing required field %q in list JSON", field)
		}
	}

	// New fields should also be present
	if _, ok := raw["backend_version"]; !ok {
		t.Error("missing new field backend_version in list JSON")
	}

	// Verify sessions array structure
	sessions, ok := raw["sessions"].([]interface{})
	if !ok {
		t.Fatal("sessions should be an array")
	}
	if len(sessions) != 1 {
		t.Fatalf("expected 1 session, got %d", len(sessions))
	}
	sess, ok := sessions[0].(map[string]interface{})
	if !ok {
		t.Fatal("session should be an object")
	}
	for _, field := range []string{"name", "clients", "started_in", "active"} {
		if _, ok := sess[field]; !ok {
			t.Errorf("missing required field sessions[0].%q in list JSON", field)
		}
	}
}
