package config

import (
	"testing"
)

func TestGetEndpointPool_MultiEndpoint(t *testing.T) {
	setEndpoints([]ApiEndpoint{
		{BaseURL: "https://a.com/v1", APIKey: "sk-a"},
		{BaseURL: "https://b.com/v1", APIKey: "sk-b"},
	}, false)
	pool := GetEndpointPool()
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
	setEndpoints([]ApiEndpoint{
		{BaseURL: "https://custom.com/v1", APIKey: "sk-x"},
	}, false)
	pool := GetEndpointPool()
	if len(pool) != 1 {
		t.Fatalf("expected 1 endpoint, got %d", len(pool))
	}
	if pool[0].BaseURL != "https://custom.com/v1" || pool[0].APIKey != "sk-x" {
		t.Errorf("endpoint mismatch: %+v", pool[0])
	}
}

func TestGetEndpointPool_Empty(t *testing.T) {
	setEndpoints(nil, false)
	pool := GetEndpointPool()
	if len(pool) != 0 {
		t.Fatalf("expected 0 endpoints, got %d", len(pool))
	}
}

func TestSetEndpoints(t *testing.T) {
	setEndpoints(nil, false)
	setEndpoints([]ApiEndpoint{
		{BaseURL: "https://new.com/v1", APIKey: "sk-new"},
	}, false)
	pool := GetEndpointPool()
	if len(pool) != 1 || pool[0].BaseURL != "https://new.com/v1" {
		t.Errorf("setEndpoints failed: %+v", pool)
	}
}

func TestGetEndpointPool_ReturnsCopy(t *testing.T) {
	setEndpoints([]ApiEndpoint{{BaseURL: "https://copy.com/v1", APIKey: "sk-copy"}}, false)
	pool := GetEndpointPool()
	pool[0].BaseURL = "https://mutated.com/v1"

	fresh := GetEndpointPool()
	if fresh[0].BaseURL != "https://copy.com/v1" {
		t.Errorf("GetEndpointPool should return a copy, got %+v", fresh)
	}
}

func TestGetEndpointPool_SortsByPriorityDescending(t *testing.T) {
	setEndpoints([]ApiEndpoint{
		{BaseURL: "https://default.com/v1", APIKey: "sk-default"},
		{BaseURL: "https://high.com/v1", APIKey: "sk-high", Priority: 100},
		{BaseURL: "https://low.com/v1", APIKey: "sk-low", Priority: 10},
	}, false)

	pool := GetEndpointPool()
	if len(pool) != 3 {
		t.Fatalf("expected 3 endpoints, got %d", len(pool))
	}
	if pool[0].BaseURL != "https://high.com/v1" || pool[1].BaseURL != "https://low.com/v1" || pool[2].BaseURL != "https://default.com/v1" {
		t.Fatalf("endpoints not sorted by priority: %+v", pool)
	}
}

func TestGetEndpointPool_PreservesOrderForEqualPriority(t *testing.T) {
	setEndpoints([]ApiEndpoint{
		{BaseURL: "https://a.com/v1", APIKey: "sk-a", Priority: 10},
		{BaseURL: "https://b.com/v1", APIKey: "sk-b", Priority: 10},
		{BaseURL: "https://c.com/v1", APIKey: "sk-c", Priority: 10},
	}, false)

	pool := GetEndpointPool()
	if pool[0].BaseURL != "https://a.com/v1" || pool[1].BaseURL != "https://b.com/v1" || pool[2].BaseURL != "https://c.com/v1" {
		t.Fatalf("equal priority order should be stable: %+v", pool)
	}
}
