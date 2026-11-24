// Package hypersender is the official Go client for the Hypersender WhatsApp API,
// SMS API, and OTP API.
//
//	client := hypersender.NewClient() // reads HYPERSENDER_* environment variables
//	res, err := client.WhatsApp.SendText(ctx, "201234567890@c.us", "Hello from Go!", nil)
//
// Every method takes a context, sends one request, and returns the API's JSON
// response as a [Result] (or []Result for list endpoints). Failed requests return
// an [*Error] with the HTTP status code and the API's message.
package hypersender

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

// Version is the SDK version, sent in the User-Agent header.
const Version = "1.0.0"

// DefaultBaseURL is the Hypersender API base URL.
const DefaultBaseURL = "https://app.hypersender.com/api"

// Result is a JSON object returned by the API.
type Result = map[string]any

// Client is a Hypersender API client. Create one with [NewClient] and reuse it;
// it is safe for concurrent use.
type Client struct {
	WhatsApp   *WhatsAppService
	SMS        *SMSService
	OTP        *OTPService
	Enterprise *EnterpriseService
}

// Option configures a [Client].
type Option func(*options)

type credentials struct{ apiKey, instanceID string }

type options struct {
	baseURL    string
	httpClient *http.Client
	whatsapp   credentials
	sms        credentials
	otp        credentials
	enterprise string
}

// WithWhatsApp sets the WhatsApp API key and instance ID.
// Defaults: HYPERSENDER_WHATSAPP_API_KEY and HYPERSENDER_WHATSAPP_INSTANCE_ID.
func WithWhatsApp(apiKey, instanceID string) Option {
	return func(o *options) { o.whatsapp = credentials{apiKey, instanceID} }
}

// WithSMS sets the SMS API key and instance ID.
// Defaults: HYPERSENDER_SMS_API_KEY and HYPERSENDER_SMS_INSTANCE_ID.
func WithSMS(apiKey, instanceID string) Option {
	return func(o *options) { o.sms = credentials{apiKey, instanceID} }
}

// WithOTP sets the OTP API key and instance ID.
// Defaults: HYPERSENDER_OTP_API_KEY and HYPERSENDER_OTP_INSTANCE_ID.
func WithOTP(apiKey, instanceID string) Option {
	return func(o *options) { o.otp = credentials{apiKey, instanceID} }
}

// WithEnterprise sets the Enterprise API key.
// Default: HYPERSENDER_ENTERPRISE_API_KEY.
func WithEnterprise(apiKey string) Option {
	return func(o *options) { o.enterprise = apiKey }
}

// WithBaseURL overrides the API base URL, for example to point at a test server.
func WithBaseURL(baseURL string) Option {
	return func(o *options) { o.baseURL = strings.TrimRight(baseURL, "/") }
}

// WithHTTPClient sets the HTTP client used for requests. The default has a 30 second timeout.
func WithHTTPClient(c *http.Client) Option {
	return func(o *options) { o.httpClient = c }
}

// NewClient returns a client configured from the environment and the given options.
// Options override environment variables.
func NewClient(opts ...Option) *Client {
	o := options{
		baseURL:    DefaultBaseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
		whatsapp:   credentials{os.Getenv("HYPERSENDER_WHATSAPP_API_KEY"), os.Getenv("HYPERSENDER_WHATSAPP_INSTANCE_ID")},
		sms:        credentials{os.Getenv("HYPERSENDER_SMS_API_KEY"), os.Getenv("HYPERSENDER_SMS_INSTANCE_ID")},
		otp:        credentials{os.Getenv("HYPERSENDER_OTP_API_KEY"), os.Getenv("HYPERSENDER_OTP_INSTANCE_ID")},
		enterprise: os.Getenv("HYPERSENDER_ENTERPRISE_API_KEY"),
	}
	for _, opt := range opts {
		opt(&o)
	}

	newService := func(path, product string, c credentials) *service {
		return &service{http: o.httpClient, baseURL: o.baseURL + path, apiKey: c.apiKey, instanceID: c.instanceID, product: product}
	}
	whatsapp := newService("/whatsapp/v2", "WHATSAPP", o.whatsapp)
	return &Client{
		WhatsApp:   &WhatsAppService{whatsapp, &ContactsService{whatsapp}},
		SMS:        &SMSService{newService("/sms/v1", "SMS", o.sms)},
		OTP:        &OTPService{newService("/otp/v2", "OTP", o.otp)},
		Enterprise: &EnterpriseService{newService("/whatsapp/v2", "ENTERPRISE", credentials{apiKey: o.enterprise})},
	}
}

// Bool returns a pointer to v, for optional boolean fields where false must be sent explicitly.
func Bool(v bool) *bool { return &v }

// Error is returned when the API responds with an HTTP 4xx or 5xx status.
type Error struct {
	StatusCode int    // HTTP status code
	Message    string // the API's "message", or the HTTP status text
	Body       []byte // the raw response body
}

func (e *Error) Error() string {
	return fmt.Sprintf("hypersender: %d %s", e.StatusCode, e.Message)
}

// Media is a file to send: a public URL ([MediaURL]) or an upload ([MediaFile]).
type Media struct {
	url      string
	filename string
	reader   io.Reader
}

// MediaURL returns media that Hypersender downloads from a public URL.
func MediaURL(url string) Media { return Media{url: url} }

// MediaFile returns media uploaded from r, for example an *os.File.
func MediaFile(filename string, r io.Reader) Media { return Media{filename: filename, reader: r} }

func (m Media) isZero() bool { return m.url == "" && m.reader == nil }

// service is one API product with its own base URL and credentials.
type service struct {
	http       *http.Client
	baseURL    string
	apiKey     string
	instanceID string
	product    string
}

type request struct {
	method     string
	path       string
	query      url.Values
	body       any    // sent as JSON
	form       any    // sent as multipart fields, with file
	fileField  string // multipart file field name
	file       Media  // multipart file
	noInstance bool   // path is not under /{instance}
}

func (s *service) send(ctx context.Context, r request, out any) error {
	if s.apiKey == "" {
		return fmt.Errorf("hypersender: missing %s API key: set HYPERSENDER_%s_API_KEY or pass it as an option", strings.ToLower(s.product), s.product)
	}
	path := r.path
	if !r.noInstance {
		if s.instanceID == "" {
			return fmt.Errorf("hypersender: missing %s instance ID: set HYPERSENDER_%s_INSTANCE_ID or pass it as an option", strings.ToLower(s.product), s.product)
		}
		path = "/" + url.PathEscape(s.instanceID) + path
	}
	u := s.baseURL + path
	if len(r.query) > 0 {
		u += "?" + r.query.Encode()
	}

	var body io.Reader
	contentType := ""
	switch {
	case r.form != nil:
		buf, ct, err := multipartBody(r.form, r.fileField, r.file)
		if err != nil {
			return err
		}
		body, contentType = buf, ct
	case r.body != nil:
		b, err := json.Marshal(r.body)
		if err != nil {
			return err
		}
		body, contentType = bytes.NewReader(b), "application/json"
	}

	req, err := http.NewRequestWithContext(ctx, r.method, u, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "hypersender-go/"+Version)
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}

	res, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		apiErr := &Error{StatusCode: res.StatusCode, Message: http.StatusText(res.StatusCode), Body: data}
		var msg struct{ Message string }
		if json.Unmarshal(data, &msg) == nil && msg.Message != "" {
			apiErr.Message = msg.Message
		}
		return apiErr
	}
	if len(bytes.TrimSpace(data)) == 0 {
		return nil
	}
	return json.Unmarshal(data, out)
}

// object sends a request that returns a JSON object.
func (s *service) object(ctx context.Context, r request) (Result, error) {
	var out Result
	err := s.send(ctx, r, &out)
	return out, err
}

// list sends a request that returns a JSON array.
func (s *service) list(ctx context.Context, r request) ([]Result, error) {
	var out []Result
	err := s.send(ctx, r, &out)
	return out, err
}

// fields flattens a struct with json tags into string values, for query strings and forms.
// Booleans become "true"/"false"; zero values tagged omitempty are dropped.
func fields(v any) (map[string]string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	if err := dec.Decode(&m); err != nil {
		return nil, err
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		if v != nil {
			out[k] = fmt.Sprint(v)
		}
	}
	return out, nil
}

// merge combines the JSON fields of several structs (nil pointers are skipped) into one object.
func merge(parts ...any) (map[string]any, error) {
	out := map[string]any{}
	for _, p := range parts {
		b, err := json.Marshal(p)
		if err != nil {
			return nil, err
		}
		dec := json.NewDecoder(bytes.NewReader(b))
		dec.UseNumber()
		var m map[string]any
		if err := dec.Decode(&m); err != nil {
			return nil, err
		}
		for k, v := range m { // a nil pointer decodes to a nil map, which adds nothing
			out[k] = v
		}
	}
	return out, nil
}

func query(v any) url.Values {
	q := url.Values{}
	if v == nil {
		return q
	}
	m, _ := fields(v) // only called with structs of strings, numbers, and bools
	for k, val := range m {
		q.Set(k, val)
	}
	return q
}

func multipartBody(form any, fileField string, file Media) (*bytes.Buffer, string, error) {
	m, err := fields(form)
	if err != nil {
		return nil, "", err
	}
	buf := &bytes.Buffer{}
	w := multipart.NewWriter(buf)
	for k, v := range m {
		// The API validates multipart booleans as 1/0.
		switch v {
		case "true":
			v = "1"
		case "false":
			v = "0"
		}
		if err := w.WriteField(k, v); err != nil {
			return nil, "", err
		}
	}
	part, err := w.CreateFormFile(fileField, file.filename)
	if err != nil {
		return nil, "", err
	}
	if _, err := io.Copy(part, file.reader); err != nil {
		return nil, "", err
	}
	if err := w.Close(); err != nil {
		return nil, "", err
	}
	return buf, w.FormDataContentType(), nil
}

var errNeedMedia = errors.New("hypersender: media is empty: use MediaURL or MediaFile")
