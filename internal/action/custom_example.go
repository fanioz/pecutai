package action

import (
	"context"
	"encoding/json"
	"fmt"
)

// CustomExampleProvider is a minimal example implementation of the Provider interface.
// It demonstrates how to implement a custom action provider for your fork.
//
// This example is intentionally simple — it returns static responses and does not
// integrate with any real external service. Use it as a starting point for your
// own implementation.
//
// Example usage:
//
//	provider := &CustomExampleProvider{name: "my-provider"}
//	teammcp.SetActionProvider(provider)
type CustomExampleProvider struct {
	name       string
	configured bool
}

// Name returns the provider name.
func (p *CustomExampleProvider) Name() string {
	if p.name != "" {
		return p.name
	}
	return "custom"
}

// Configured reports whether the provider is configured and ready to use.
func (p *CustomExampleProvider) Configured() bool {
	return p.configured
}

// Supports reports whether the provider supports a given capability.
func (p *CustomExampleProvider) Supports(cap Capability) bool {
	// This example supports all capabilities. A real implementation would
	// check against the provider's actual capabilities.
	switch cap {
	case CapabilityGuide,
		CapabilityConnections,
		CapabilityActionSearch,
		CapabilityActionKnowledge,
		CapabilityActionExecute,
		CapabilityWorkflowCreate,
		CapabilityWorkflowExecute,
		CapabilityWorkflowRuns,
		CapabilityRelayList,
		CapabilityRelayEventTypes,
		CapabilityRelayCreate,
		CapabilityRelayActivate,
		CapabilityRelayEvents,
		CapabilityRelayEvent:
		return true
	default:
		return false
	}
}

// Guide returns the provider's usage guide.
func (p *CustomExampleProvider) Guide(ctx context.Context, topic string) (GuideResult, error) {
	guide := fmt.Sprintf("# %s Provider Guide\n\nThis is a custom action provider example.", p.Name())
	return GuideResult{
		Topic: topic,
		Guide: guide,
		Raw:   json.RawMessage(fmt.Sprintf(`{"topic": %q, "guide": %q}`, topic, guide)),
	}, nil
}

// ListConnections returns the provider's connected accounts.
func (p *CustomExampleProvider) ListConnections(ctx context.Context, opts ListConnectionsOptions) (ConnectionsResult, error) {
	return ConnectionsResult{
		Connections: []Connection{
			{
				Platform: "example",
				State:    "connected",
				Key:      "example-connection-1",
				Name:     "Example Account",
			},
		},
	}, nil
}

// SearchActions searches for actions on a platform using natural language.
func (p *CustomExampleProvider) SearchActions(ctx context.Context, platform, query, mode string) (ActionSearchResult, error) {
	return ActionSearchResult{
		Platform: platform,
		Query:    query,
		Mode:     mode,
		Actions: []Action{
			{
				ActionID: "EXAMPLE_SEND_MESSAGE",
				Title:    "Send Message",
				Method:   "POST",
				Path:     "/api/v1/messages",
			},
		},
	}, nil
}

// ActionKnowledge returns the schema and usage guidance for an action.
func (p *CustomExampleProvider) ActionKnowledge(ctx context.Context, platform, actionID string) (KnowledgeResult, error) {
	knowledge := fmt.Sprintf("Action: %s on %s\n\nThis is an example action.", actionID, platform)
	return KnowledgeResult{
		Platform:  platform,
		ActionID:  actionID,
		Knowledge: knowledge,
	}, nil
}

// ExecuteAction executes an external action.
func (p *CustomExampleProvider) ExecuteAction(ctx context.Context, req ExecuteRequest) (ExecuteResult, error) {
	return ExecuteResult{
		DryRun: req.DryRun,
		Request: ExecuteEnvelope{
			Method: "POST",
			URL:    fmt.Sprintf("https://api.example.com/actions/%s", req.ActionID),
			Headers: req.Headers,
			Data:    req.Data,
		},
		Response: json.RawMessage(`{"status": "success", "message": "This is a mock response from the custom provider."}`),
	}, nil
}

// CreateWorkflow creates a reusable workflow.
func (p *CustomExampleProvider) CreateWorkflow(ctx context.Context, req WorkflowCreateRequest) (WorkflowCreateResult, error) {
	return WorkflowCreateResult{
		Created: true,
		Key:     req.Key,
		Path:    "/workflows/" + req.Key,
	}, nil
}

// ExecuteWorkflow executes a saved workflow.
func (p *CustomExampleProvider) ExecuteWorkflow(ctx context.Context, req WorkflowExecuteRequest) (WorkflowExecuteResult, error) {
	return WorkflowExecuteResult{
		RunID:  "example-run-123",
		Status: "completed",
	}, nil
}

// ListWorkflowRuns returns runs for a workflow.
func (p *CustomExampleProvider) ListWorkflowRuns(ctx context.Context, key string) (WorkflowRunsResult, error) {
	return WorkflowRunsResult{
		Runs: []json.RawMessage{
			json.RawMessage(`{"run_id": "example-run-123", "status": "completed"}`),
		},
	}, nil
}

// ListRelays returns registered relay endpoints.
func (p *CustomExampleProvider) ListRelays(ctx context.Context, opts ListRelaysOptions) (RelayListResult, error) {
	return RelayListResult{
		Total: 0,
		Endpoints: []Relay{},
	}, nil
}

// RelayEventTypes returns supported event types for a platform.
func (p *CustomExampleProvider) RelayEventTypes(ctx context.Context, platform string) (RelayEventTypesResult, error) {
	return RelayEventTypesResult{
		Platform:   platform,
		EventTypes: []string{},
	}, nil
}

// CreateRelay creates a new relay endpoint.
func (p *CustomExampleProvider) CreateRelay(ctx context.Context, req RelayCreateRequest) (RelayResult, error) {
	return RelayResult{
		ID: "example-relay-123",
	}, nil
}

// ActivateRelay activates a relay endpoint.
func (p *CustomExampleProvider) ActivateRelay(ctx context.Context, req RelayActivateRequest) (RelayResult, error) {
	return RelayResult{
		ID:     req.ID,
		Active: true,
	}, nil
}

// ListRelayEvents lists recent relay events.
func (p *CustomExampleProvider) ListRelayEvents(ctx context.Context, opts RelayEventsOptions) (RelayEventsResult, error) {
	return RelayEventsResult{
		Total:   0,
		Showing: 0,
		Events:  []RelayEvent{},
	}, nil
}

// GetRelayEvent fetches a single relay event.
func (p *CustomExampleProvider) GetRelayEvent(ctx context.Context, id string) (RelayEventDetail, error) {
	return RelayEventDetail{
		ID: id,
		Payload: json.RawMessage(fmt.Sprintf(`{"event_id": %q, "message": "example payload"}`, id)),
	}, nil
}
