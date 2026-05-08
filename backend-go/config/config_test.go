package config

import (
	"encoding/json"
	"testing"
)

func TestGetEndpointPool_MultiEndpoint(t *testing.T) {
	c := &Config{
		ApiEndpoints: []ApiEndpoint{
			{BaseURL: "https://a.com/v1", APIKey: "sk-a"},
			{BaseURL: "https://b.com/v1", APIKey: "sk-b"},
		},
	}
	pool := c.GetEndpointPool()
	if len(pool) != 2 {
		t.Fatalf("expected 2 endpoints, got %d", len(pool))
	}
	if pool[0].BaseURL != "https://a.com/v1" || pool[0].APIKey != "sk-a" {
		t.Errorf("endpoint 0 mismatch: %+v", pool[0])
	}
	if pool[1].BaseURL != "https://b.com/v1" || pool[1].APIKey != "sk-b" {
		t.Errorf("endpoint 1 mismatch: %+v", pool[1])
	}
}

func TestGetEndpointPool_SingleEndpoint(t *testing.T) {
	c := &Config{
		ApiEndpoints: []ApiEndpoint{
			{BaseURL: "https://custom.com/v1", APIKey: "sk-x"},
		},
	}
	pool := c.GetEndpointPool()
	if len(pool) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(pool))
	}
	if pool[0].BaseURL != "https://custom.com/v1" || pool[0].APIKey != "sk-x" {
		t.Errorf("endpoint mismatch: %+v", pool[0])
	}
}

func TestLoad_WithApiEndpoints(t *testing.T) {
	c := &Config{}
	data := []byte(`{"defaults":{"codexCli":true},"apiEndpoints":[{"baseUrl":"https://a.com/v1","apiKey":"sk-a"}]}`)
	if err := json.Unmarshal(data, c); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(c.ApiEndpoints) != 1 {
		t.Fatalf("expected 1 apiEndpoint, got %d", len(c.ApiEndpoints))
	}
	if c.ApiEndpoints[0].BaseURL != "https://a.com/v1" {
		t.Errorf("expected base URL https://a.com/v1, got %s", c.ApiEndpoints[0].BaseURL)
	}
}

func TestLoad_NoDefaultsBaseURL(t *testing.T) {
	// Verify that JSON containing "baseUrl" inside defaults does not cause
	// an error — the field is simply ignored since Defaults no longer has BaseURL.
	c := &Config{}
	data := []byte(`{"defaults":{"codexCli":true,"baseUrl":"https://ignored.com/v1","model":"gpt-image-2"},"apiEndpoints":[{"baseUrl":"https://a.com/v1","apiKey":"sk-a"}]}`)
	if err := json.Unmarshal(data, c); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if c.Defaults.Model != "gpt-image-2" {
		t.Errorf("expected model gpt-image-2, got %s", c.Defaults.Model)
	}
	if !c.Defaults.CodexCLI {
		t.Errorf("expected codexCli true, got false")
	}
	// apiEndpoints should still be parsed correctly
	if len(c.ApiEndpoints) != 1 || c.ApiEndpoints[0].BaseURL != "https://a.com/v1" {
		t.Errorf("apiEndpoints not parsed correctly: %+v", c.ApiEndpoints)
	}
}
