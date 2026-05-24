package service

import (
	"errors"
	"testing"

	"gpt-image-playground/backend/config"
)

func TestWithFailoverStampsEndpointAttribution(t *testing.T) {
	endpoints := []config.ApiEndpoint{
		{BaseURL: "https://success.example/v1", APIKey: "sk-success", Priority: 1, MaxConcurrency: 2, CostPerImageX10000: 2345},
	}

	// fn succeeds immediately — attribution should be stamped
	fn := func(apiKey, baseURL string) (*ImageGenResult, error) {
		return &ImageGenResult{
			Images: []GeneratedImage{
				{Base64: "data:image/png;base64,aaa", RevisedPrompt: "test prompt"},
			},
		}, nil
	}

	result, err := withFailover(endpoints, "", nil, fn)
	if err != nil {
		t.Fatalf("withFailover: %v", err)
	}

	if len(result.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(result.Images))
	}

	img := result.Images[0]
	if img.EndpointBaseURL != "https://success.example/v1" {
		t.Errorf("EndpointBaseURL = %q, want https://success.example/v1", img.EndpointBaseURL)
	}
	if img.UnitCostX10000 != 2345 {
		t.Errorf("UnitCostX10000 = %d, want 2345", img.UnitCostX10000)
	}
}

func TestWithFailoverFailedFirstEndpointDoesNotStampAttribution(t *testing.T) {
	callCount := 0

	endpoints := []config.ApiEndpoint{
		{BaseURL: "https://fail.example/v1", APIKey: "sk-fail", Priority: 2, CostPerImageX10000: 9999},
		{BaseURL: "https://success.example/v1", APIKey: "sk-success", Priority: 1, CostPerImageX10000: 2345},
	}

	fn := func(apiKey, baseURL string) (*ImageGenResult, error) {
		callCount++
		if baseURL == "https://fail.example/v1" {
			return nil, errors.New("simulated endpoint failure")
		}
		return &ImageGenResult{
			Images: []GeneratedImage{
				{Base64: "data:image/png;base64,bbb", RevisedPrompt: "from success"},
			},
		}, nil
	}

	result, err := withFailover(endpoints, "", nil, fn)
	if err != nil {
		t.Fatalf("withFailover: %v", err)
	}

	if callCount < 2 {
		t.Fatalf("expected at least 2 fn calls (failover happened), got %d", callCount)
	}

	if len(result.Images) != 1 {
		t.Fatalf("expected 1 image, got %d", len(result.Images))
	}

	img := result.Images[0]
	// Attribution must come from the SUCCESSFUL endpoint, not the failed one
	if img.EndpointBaseURL != "https://success.example/v1" {
		t.Errorf("EndpointBaseURL = %q, want https://success.example/v1 (should NOT be fail URL)", img.EndpointBaseURL)
	}
	if img.UnitCostX10000 != 2345 {
		t.Errorf("UnitCostX10000 = %d, want 2345 (should NOT be 9999 from failed endpoint)", img.UnitCostX10000)
	}
}

func TestMergeConcurrentResultsPreservesEndpointAttribution(t *testing.T) {
	results := []*ImageGenResult{
		{
			Images: []GeneratedImage{
				{Base64: "data:image/png;base64,c1", EndpointBaseURL: "https://ep1.example/v1", UnitCostX10000: 1000},
				{Base64: "data:image/png;base64,c2", EndpointBaseURL: "https://ep1.example/v1", UnitCostX10000: 1000},
			},
			ActualParams: map[string]interface{}{"size": "1024x1024"},
		},
		{
			Images: []GeneratedImage{
				{Base64: "data:image/png;base64,c3", EndpointBaseURL: "https://ep2.example/v1", UnitCostX10000: 2000},
			},
		},
	}
	errs := []error{nil, nil}

	merged, err := mergeConcurrentResults(results, errs)
	if err != nil {
		t.Fatalf("mergeConcurrentResults: %v", err)
	}

	if len(merged.Images) != 3 {
		t.Fatalf("expected 3 merged images, got %d", len(merged.Images))
	}

	// Check that endpoint attribution is preserved per-image
	if merged.Images[0].EndpointBaseURL != "https://ep1.example/v1" {
		t.Errorf("image 0: EndpointBaseURL = %q, want https://ep1.example/v1", merged.Images[0].EndpointBaseURL)
	}
	if merged.Images[0].UnitCostX10000 != 1000 {
		t.Errorf("image 0: UnitCostX10000 = %d, want 1000", merged.Images[0].UnitCostX10000)
	}
	if merged.Images[2].EndpointBaseURL != "https://ep2.example/v1" {
		t.Errorf("image 2: EndpointBaseURL = %q, want https://ep2.example/v1", merged.Images[2].EndpointBaseURL)
	}
	if merged.Images[2].UnitCostX10000 != 2000 {
		t.Errorf("image 2: UnitCostX10000 = %d, want 2000", merged.Images[2].UnitCostX10000)
	}

	// Also verify actual params got merged
	if merged.ActualParams["n"] != 3 {
		t.Errorf("ActualParams[n] = %v, want 3", merged.ActualParams["n"])
	}
}
