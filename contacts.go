package hypersender

import (
	"context"
	"net/http"
	"net/url"
)

// ContactsService looks up and saves WhatsApp contacts. Use it as client.WhatsApp.Contacts.
type ContactsService struct{ s *service }

// CheckContactExists checks if a phone number has WhatsApp. phone has the country
// code, without + or spaces, e.g. "201234567890". The result has numberExists and chatId.
func (c *ContactsService) CheckContactExists(ctx context.Context, phone string) (Result, error) {
	body := struct {
		Phone string `json:"phone"`
	}{phone}
	return c.s.object(ctx, request{method: http.MethodPost, path: "/contacts/check-exists", body: body})
}

// GetAllContactsOptions are the optional fields for GetAllContacts.
type GetAllContactsOptions struct {
	Limit     int    `json:"limit,omitempty"`      // 1-1000, default 100
	Offset    int    `json:"offset,omitempty"`     // default 0
	SortBy    string `json:"sort_by,omitempty"`    // "id" or "name"
	SortOrder string `json:"sort_order,omitempty"` // "asc" or "desc"
}

// GetAllContacts lists the instance's contacts.
func (c *ContactsService) GetAllContacts(ctx context.Context, opts *GetAllContactsOptions) (Result, error) {
	return c.s.object(ctx, request{method: http.MethodGet, path: "/contacts/get-all", query: query(opts)})
}

// GetContact returns a contact by phone number or chat ID ("123@c.us" or "123@lid").
func (c *ContactsService) GetContact(ctx context.Context, chatID string) (Result, error) {
	return c.s.object(ctx, request{method: http.MethodGet, path: "/contacts/get", query: url.Values{"chat_id": {chatID}}})
}

// GetProfilePicture returns the URL of a contact's WhatsApp profile picture.
func (c *ContactsService) GetProfilePicture(ctx context.Context, chatID string) (Result, error) {
	return c.s.object(ctx, request{method: http.MethodGet, path: "/contacts/profile-picture", query: url.Values{"chat_id": {chatID}}})
}

// UpsertContactOptions are the optional fields for UpsertContact.
type UpsertContactOptions struct {
	LastName string `json:"last_name,omitempty"`
}

// UpsertContact creates a contact, or updates its name if it already exists.
func (c *ContactsService) UpsertContact(ctx context.Context, chatID, firstName string, opts *UpsertContactOptions) (Result, error) {
	body := struct {
		ChatID    string `json:"chat_id"`
		FirstName string `json:"first_name"`
		*UpsertContactOptions
	}{chatID, firstName, opts}
	return c.s.object(ctx, request{method: http.MethodPost, path: "/contacts/upsert", body: body})
}

// GetAllKnownLIDsOptions are the optional fields for GetAllKnownLIDs.
type GetAllKnownLIDsOptions struct {
	Limit  int `json:"limit,omitempty"`  // default 100
	Offset int `json:"offset,omitempty"` // default 0
}

// GetAllKnownLIDs lists the LID to phone number mappings the instance knows.
func (c *ContactsService) GetAllKnownLIDs(ctx context.Context, opts *GetAllKnownLIDsOptions) ([]Result, error) {
	return c.s.list(ctx, request{method: http.MethodGet, path: "/contacts/get-all-known-lids", query: query(opts)})
}

// GetLIDsCount returns how many LID mappings the instance knows.
func (c *ContactsService) GetLIDsCount(ctx context.Context) (Result, error) {
	return c.s.object(ctx, request{method: http.MethodGet, path: "/contacts/get-lids-count"})
}

// GetPhoneNumberByLID resolves a LID ("1111111@lid") to its phone number chat ID.
func (c *ContactsService) GetPhoneNumberByLID(ctx context.Context, lid string) (Result, error) {
	return c.s.object(ctx, request{method: http.MethodGet, path: "/contacts/get-phone-number-by-lid", query: url.Values{"lid": {lid}}})
}

// GetLIDByPhoneNumber finds the LID linked to a phone number ("3333333@c.us" or "3333333").
func (c *ContactsService) GetLIDByPhoneNumber(ctx context.Context, phoneNumber string) (Result, error) {
	return c.s.object(ctx, request{method: http.MethodGet, path: "/contacts/get-lid-by-phone-number", query: url.Values{"phoneNumber": {phoneNumber}}})
}
