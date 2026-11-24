package hypersender

import (
	"context"
	"net/http"
)

// OTPService sends and checks one-time passwords over WhatsApp or SMS. Use it as client.OTP.
type OTPService struct{ s *service }

// RequestCodeOptions are the optional fields for RequestCode.
type RequestCodeOptions struct {
	Length     int    `json:"length,omitempty"`     // 4-10
	UseNumber  *bool  `json:"useNumber,omitempty"`  // use hypersender.Bool(true)
	UseLetter  *bool  `json:"useLetter,omitempty"`  // use hypersender.Bool(true)
	AllCapital *bool  `json:"allCapital,omitempty"` // use hypersender.Bool(true)
	Name       string `json:"name,omitempty"`       // your app or company name, shown in the message
	Expires    int    `json:"expires,omitempty"`    // seconds, 60-1800
}

// RequestCode sends a one-time code to chatID.
func (o *OTPService) RequestCode(ctx context.Context, chatID string, opts *RequestCodeOptions) (Result, error) {
	body := struct {
		ChatID string `json:"chatId"`
		*RequestCodeOptions
	}{chatID, opts}
	return o.s.object(ctx, request{method: http.MethodPost, path: "/request-code", body: body})
}

// ValidateCode checks a code the user entered.
func (o *OTPService) ValidateCode(ctx context.Context, chatID, code string) (Result, error) {
	body := struct {
		ChatID string `json:"chatId"`
		Code   string `json:"code"`
	}{chatID, code}
	return o.s.object(ctx, request{method: http.MethodPost, path: "/validate-code", body: body})
}

// OTPLinkMessage overrides the messages shown during OTP link verification.
type OTPLinkMessage struct {
	Prompt  string `json:"prompt,omitempty"`
	Success string `json:"success,omitempty"`
	Failed  string `json:"failed,omitempty"`
	Expired string `json:"expired,omitempty"`
}

// OTPLinkCallback sets where the user is sent after verification.
type OTPLinkCallback struct {
	Success string `json:"success,omitempty"`
	Failed  string `json:"failed,omitempty"`
}

// GenerateLinkOptions are the optional fields for GenerateLink.
type GenerateLinkOptions struct {
	Expires  int              `json:"expires,omitempty"` // seconds, 60-3600
	Name     string           `json:"name,omitempty"`
	Message  *OTPLinkMessage  `json:"message,omitempty"`
	Callback *OTPLinkCallback `json:"callback,omitempty"`
}

// GenerateLink creates a WhatsApp OTP link the user opens to verify. The link is in data.whatsapp_link.
func (o *OTPService) GenerateLink(ctx context.Context, chatID string, opts *GenerateLinkOptions) (Result, error) {
	body := struct {
		ChatID string `json:"chatId"`
		*GenerateLinkOptions
	}{chatID, opts}
	return o.s.object(ctx, request{method: http.MethodPost, path: "/generate-link", body: body})
}
