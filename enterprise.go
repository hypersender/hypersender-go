package hypersender

import (
	"context"
	"net/http"
	"net/url"
)

// EnterpriseService manages WhatsApp instances on an Enterprise account. Use it as
// client.Enterprise. The instance ID is passed on each call instead of being set on the client.
type EnterpriseService struct{ s *service }

// ListInstancesOptions filter ListInstances.
type ListInstancesOptions struct {
	UUID   string `json:"filter[uuid],omitempty"`
	Name   string `json:"filter[name],omitempty"` // partial match
	Phone  string `json:"filter[phone],omitempty"`
	Status string `json:"filter[status],omitempty"` // STOPPED, STARTING, WORKING, SCAN_QR_CODE, FAILED, PENDING
}

// ListInstances lists your WhatsApp instances.
func (e *EnterpriseService) ListInstances(ctx context.Context, opts *ListInstancesOptions) ([]Result, error) {
	return e.s.list(ctx, request{method: http.MethodGet, path: "/instances", query: query(opts), noInstance: true})
}

// CreateInstanceOptions are the optional fields for CreateInstance.
type CreateInstanceOptions struct {
	WebhookURL string `json:"webhook_url,omitempty"` // an https URL
}

// CreateInstance creates a WhatsApp instance. phone is E.164 ("+201234567890").
// The result has the instance id and a qr_code to scan with the phone.
func (e *EnterpriseService) CreateInstance(ctx context.Context, phone, name string, opts *CreateInstanceOptions) (Result, error) {
	body := struct {
		Phone string `json:"phone"`
		Name  string `json:"name"`
		*CreateInstanceOptions
	}{phone, name, opts}
	return e.s.object(ctx, request{method: http.MethodPost, path: "/create-instance", body: body, noInstance: true})
}

// GetInstanceHealth checks whether an instance is connected and working.
func (e *EnterpriseService) GetInstanceHealth(ctx context.Context, instanceID string) (Result, error) {
	return e.s.object(ctx, request{method: http.MethodGet, path: "/" + url.PathEscape(instanceID) + "/health", noInstance: true})
}

// RequestNewQRCode returns a fresh QR code to connect the phone.
func (e *EnterpriseService) RequestNewQRCode(ctx context.Context, instanceID string) (Result, error) {
	return e.s.object(ctx, request{method: http.MethodPost, path: "/" + url.PathEscape(instanceID) + "/request-new-qr-code", noInstance: true})
}

// RestartInstance restarts an instance.
func (e *EnterpriseService) RestartInstance(ctx context.Context, instanceID string) (Result, error) {
	return e.s.object(ctx, request{method: http.MethodPost, path: "/" + url.PathEscape(instanceID) + "/restart", noInstance: true})
}
