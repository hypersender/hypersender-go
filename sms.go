package hypersender

import (
	"context"
	"net/http"
	"net/url"
)

// SMSService sends SMS through your Android phone. Use it as client.SMS.
type SMSService struct{ s *service }

// SendSMSOptions are the optional fields for SendMessage.
type SendSMSOptions struct {
	RequestID                string `json:"request_id,omitempty"`                 // your own ID, echoed back in webhooks
	ScheduleSendAt           string `json:"schedule_send_at,omitempty"`           // send later, e.g. "2026-12-01 09:00:00"
	MaxSendAttempts          int    `json:"max_send_attempts,omitempty"`          // retries on failure
	MessageExpirationSeconds int    `json:"message_expiration_seconds,omitempty"` // give up after this many seconds
}

// SendMessage sends an SMS to a phone number such as "+201234567890".
func (s *SMSService) SendMessage(ctx context.Context, to, content string, opts *SendSMSOptions) (Result, error) {
	body := struct {
		To      string `json:"to"`
		Content string `json:"content"`
		*SendSMSOptions
	}{to, content, opts}
	return s.s.object(ctx, request{method: http.MethodPost, path: "/send-message", body: body})
}

// ListOptions filter, sort, and page SMS lists.
type ListOptions struct {
	Filters map[string]string // e.g. {"status": "delivered"} becomes filter[status]=delivered
	Sort    string            // e.g. "-created_at"
	Include string
	Page    int
	PerPage int
}

func (o *ListOptions) values() url.Values {
	if o == nil {
		return nil
	}
	q := query(struct {
		Sort    string `json:"sort,omitempty"`
		Include string `json:"include,omitempty"`
		Page    int    `json:"page,omitempty"`
		PerPage int    `json:"per_page,omitempty"`
	}{o.Sort, o.Include, o.Page, o.PerPage})
	for k, v := range o.Filters {
		q.Set("filter["+k+"]", v)
	}
	return q
}

// ListMessages lists messages.
func (s *SMSService) ListMessages(ctx context.Context, opts *ListOptions) (Result, error) {
	return s.s.object(ctx, request{method: http.MethodGet, path: "/messages", query: opts.values()})
}

// DeleteMessage deletes a message.
func (s *SMSService) DeleteMessage(ctx context.Context, messageID string) (Result, error) {
	return s.s.object(ctx, request{method: http.MethodDelete, path: "/delete-message", query: url.Values{"messageId": {messageID}}})
}

// ListMessageThreads lists conversations.
func (s *SMSService) ListMessageThreads(ctx context.Context, opts *ListOptions) (Result, error) {
	return s.s.object(ctx, request{method: http.MethodGet, path: "/message-threads", query: opts.values()})
}

// EditMessageThread archives or unarchives a conversation.
func (s *SMSService) EditMessageThread(ctx context.Context, messageThreadID string, isArchived bool) (Result, error) {
	body := struct {
		MessageThreadID string `json:"messageThreadId"`
		IsArchived      bool   `json:"is_archived"`
	}{messageThreadID, isArchived}
	return s.s.object(ctx, request{method: http.MethodPut, path: "/edit-message-thread", body: body})
}

// DeleteMessageThread deletes a conversation.
func (s *SMSService) DeleteMessageThread(ctx context.Context, messageThreadID string) (Result, error) {
	return s.s.object(ctx, request{method: http.MethodDelete, path: "/delete-message-thread", query: url.Values{"messageThreadId": {messageThreadID}}})
}

// GetHeartbeats lists the phone's heartbeats (app version, battery, charging) to check it is online.
func (s *SMSService) GetHeartbeats(ctx context.Context, opts *ListOptions) (Result, error) {
	return s.s.object(ctx, request{method: http.MethodGet, path: "/heartbeats", query: opts.values()})
}
