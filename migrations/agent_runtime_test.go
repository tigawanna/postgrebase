package migrations

import "testing"

func TestAgentRuntimeMySQLTypes(t *testing.T) {
	if got := agentStringType("mysql"); got != "VARCHAR(255)" {
		t.Fatalf("expected mysql agent string type VARCHAR(255), got %q", got)
	}
	if got := agentLongTextType("mysql"); got != "LONGTEXT" {
		t.Fatalf("expected mysql agent long text type LONGTEXT, got %q", got)
	}
	if got := agentJsonType("mysql"); got != "JSON" {
		t.Fatalf("expected mysql agent json type JSON, got %q", got)
	}
}
