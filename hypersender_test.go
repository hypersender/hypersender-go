package hypersender

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

type captured struct {
	method, url, auth, contentType string
	body                           []byte
}

// newTestClient returns a client pointed at a test server that records each request
// and answers with status and response.
func newTestClient(t *testing.T, status int, response string) (*Client, *[]captured) {
	t.Helper()
	var got []captured
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		got = append(got, captured{r.Method, r.URL.RequestURI(), r.Header.Get("Authorization"), r.Header.Get("Content-Type"), body})
		w.WriteHeader(status)
		io.WriteString(w, response)
	}))
	t.Cleanup(srv.Close)
	return NewClient(
		WithBaseURL(srv.URL),
		WithWhatsApp("wa-key", "wa-1"),
		WithSMS("sms-key", "sms-1"),
		WithOTP("otp-key", "otp-1"),
		WithEnterprise("ent-key"),
	), &got
}

func TestEndpoints(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name   string
		call   func(c *Client) error
		method string
		url    string
		auth   string
		body   string // expected JSON body, "" for none
	}{
		// WhatsApp: sending
		{"SendText", func(c *Client) error { _, err := c.WhatsApp.SendText(ctx, "1@c.us", "Hi", nil); return err },
			"POST", "/whatsapp/v2/wa-1/send-text", "wa-key", `{"chatId":"1@c.us","text":"Hi"}`},
		{"SendText options", func(c *Client) error {
			_, err := c.WhatsApp.SendText(ctx, "1@c.us", "Hi", &SendTextOptions{ReplyTo: "m1", LinkPreview: true})
			return err
		}, "POST", "/whatsapp/v2/wa-1/send-text", "wa-key", `{"chatId":"1@c.us","text":"Hi","reply_to":"m1","link_preview":true}`},
		{"SendTextSafe", func(c *Client) error { _, err := c.WhatsApp.SendTextSafe(ctx, "1@c.us", "Hi", nil); return err },
			"POST", "/whatsapp/v2/wa-1/send-text-safe", "wa-key", `{"chatId":"1@c.us","text":"Hi"}`},
		{"SendLinkCustomPreview", func(c *Client) error {
			_, err := c.WhatsApp.SendLinkCustomPreview(ctx, "1@c.us", "Look", LinkPreview{Title: "T", Description: "D", URL: "https://x.com", Image: MediaURL("https://x.com/i.png")}, nil)
			return err
		}, "POST", "/whatsapp/v2/wa-1/send-link-custom-preview", "wa-key",
			`{"chatId":"1@c.us","text":"Look","preview_title":"T","preview_description":"D","preview_url":"https://x.com","preview_image_url":"https://x.com/i.png"}`},
		{"SendFile", func(c *Client) error {
			_, err := c.WhatsApp.SendFile(ctx, "1@c.us", MediaURL("https://x.com/a.pdf"), &MediaOptions{Caption: "Doc"})
			return err
		}, "POST", "/whatsapp/v2/wa-1/send-file", "wa-key", `{"chatId":"1@c.us","url":"https://x.com/a.pdf","caption":"Doc"}`},
		{"SendImage", func(c *Client) error {
			_, err := c.WhatsApp.SendImage(ctx, "1@c.us", MediaURL("https://x.com/a.png"), nil)
			return err
		},
			"POST", "/whatsapp/v2/wa-1/send-image", "wa-key", `{"chatId":"1@c.us","url":"https://x.com/a.png"}`},
		{"SendVideo", func(c *Client) error {
			_, err := c.WhatsApp.SendVideo(ctx, "1@c.us", MediaURL("https://x.com/a.mp4"), nil)
			return err
		},
			"POST", "/whatsapp/v2/wa-1/send-video", "wa-key", `{"chatId":"1@c.us","url":"https://x.com/a.mp4"}`},
		{"SendVoice", func(c *Client) error {
			_, err := c.WhatsApp.SendVoice(ctx, "1@c.us", MediaURL("https://x.com/a.ogg"), nil)
			return err
		},
			"POST", "/whatsapp/v2/wa-1/send-voice", "wa-key", `{"chatId":"1@c.us","url":"https://x.com/a.ogg"}`},
		{"SendPoll", func(c *Client) error {
			_, err := c.WhatsApp.SendPoll(ctx, "1@c.us", "Lunch?", []string{"Yes", "No"}, nil)
			return err
		},
			"POST", "/whatsapp/v2/wa-1/send-poll", "wa-key", `{"chatId":"1@c.us","poll":{"name":"Lunch?","options":["Yes","No"],"multipleAnswers":false}}`},
		{"SendContactCard", func(c *Client) error {
			_, err := c.WhatsApp.SendContactCard(ctx, "1@c.us", []ContactCard{{FullName: "A", PhoneNumber: "2"}})
			return err
		}, "POST", "/whatsapp/v2/wa-1/send-contact-card", "wa-key", `{"chatId":"1@c.us","contacts":[{"fullName":"A","phoneNumber":"2"}]}`},
		{"SendLocation", func(c *Client) error {
			_, err := c.WhatsApp.SendLocation(ctx, "1@c.us", 30.04, 31.23, &SendLocationOptions{Title: "Cairo"})
			return err
		}, "POST", "/whatsapp/v2/wa-1/send-location", "wa-key", `{"chatId":"1@c.us","latitude":30.04,"longitude":31.23,"title":"Cairo"}`},
		{"ForwardMessage", func(c *Client) error { _, err := c.WhatsApp.ForwardMessage(ctx, "1@c.us", "m1"); return err },
			"POST", "/whatsapp/v2/wa-1/forward-message", "wa-key", `{"chatId":"1@c.us","messageId":"m1"}`},
		// WhatsApp: messages and chats
		{"ReactToMessage", func(c *Client) error { _, err := c.WhatsApp.ReactToMessage(ctx, "1@c.us", "m1", "👍"); return err },
			"PUT", "/whatsapp/v2/wa-1/react-to-message", "wa-key", `{"chatId":"1@c.us","messageId":"m1","reaction":"👍"}`},
		{"StarMessage", func(c *Client) error { _, err := c.WhatsApp.StarMessage(ctx, "1@c.us", "m1", false); return err },
			"PUT", "/whatsapp/v2/wa-1/star-message", "wa-key", `{"chatId":"1@c.us","messageId":"m1","star":false}`},
		{"DeleteMessage", func(c *Client) error { _, err := c.WhatsApp.DeleteMessage(ctx, "1@c.us", "m1"); return err },
			"DELETE", "/whatsapp/v2/wa-1/delete-message?chatId=1%40c.us&messageId=m1", "wa-key", ""},
		{"GetQueuedRequest", func(c *Client) error { _, err := c.WhatsApp.GetQueuedRequest(ctx, "q-1"); return err },
			"GET", "/whatsapp/v2/wa-1/queued-requests/q-1", "wa-key", ""},
		{"SendSeen", func(c *Client) error { _, err := c.WhatsApp.SendSeen(ctx, "1@c.us", "m1", nil); return err },
			"POST", "/whatsapp/v2/wa-1/send-seen", "wa-key", `{"chatId":"1@c.us","messageId":"m1"}`},
		{"ReadChat", func(c *Client) error {
			_, err := c.WhatsApp.ReadChat(ctx, "1@c.us", &ReadChatOptions{Days: 7})
			return err
		},
			"POST", "/whatsapp/v2/wa-1/read-chat?days=7", "wa-key", `{"chatId":"1@c.us"}`},
		{"StartTyping", func(c *Client) error { _, err := c.WhatsApp.StartTyping(ctx, "1@c.us"); return err },
			"POST", "/whatsapp/v2/wa-1/start-typing", "wa-key", `{"chatId":"1@c.us"}`},
		{"StopTyping", func(c *Client) error { _, err := c.WhatsApp.StopTyping(ctx, "1@c.us"); return err },
			"POST", "/whatsapp/v2/wa-1/stop-typing", "wa-key", `{"chatId":"1@c.us"}`},
		// WhatsApp: contacts
		{"CheckContactExists", func(c *Client) error {
			_, err := c.WhatsApp.Contacts.CheckContactExists(ctx, "201234567890")
			return err
		},
			"POST", "/whatsapp/v2/wa-1/contacts/check-exists", "wa-key", `{"phone":"201234567890"}`},
		{"GetAllContacts", func(c *Client) error {
			_, err := c.WhatsApp.Contacts.GetAllContacts(ctx, &GetAllContactsOptions{Limit: 50, SortBy: "name"})
			return err
		}, "GET", "/whatsapp/v2/wa-1/contacts/get-all?limit=50&sort_by=name", "wa-key", ""},
		{"GetAllContacts nil", func(c *Client) error { _, err := c.WhatsApp.Contacts.GetAllContacts(ctx, nil); return err },
			"GET", "/whatsapp/v2/wa-1/contacts/get-all", "wa-key", ""},
		{"GetContact", func(c *Client) error { _, err := c.WhatsApp.Contacts.GetContact(ctx, "1@lid"); return err },
			"GET", "/whatsapp/v2/wa-1/contacts/get?chat_id=1%40lid", "wa-key", ""},
		{"GetProfilePicture", func(c *Client) error { _, err := c.WhatsApp.Contacts.GetProfilePicture(ctx, "1@c.us"); return err },
			"GET", "/whatsapp/v2/wa-1/contacts/profile-picture?chat_id=1%40c.us", "wa-key", ""},
		{"UpsertContact", func(c *Client) error {
			_, err := c.WhatsApp.Contacts.UpsertContact(ctx, "1@c.us", "John", nil)
			return err
		},
			"POST", "/whatsapp/v2/wa-1/contacts/upsert", "wa-key", `{"chat_id":"1@c.us","first_name":"John"}`},
		{"GetLIDsCount", func(c *Client) error { _, err := c.WhatsApp.Contacts.GetLIDsCount(ctx); return err },
			"GET", "/whatsapp/v2/wa-1/contacts/get-lids-count", "wa-key", ""},
		{"GetPhoneNumberByLID", func(c *Client) error { _, err := c.WhatsApp.Contacts.GetPhoneNumberByLID(ctx, "1@lid"); return err },
			"GET", "/whatsapp/v2/wa-1/contacts/get-phone-number-by-lid?lid=1%40lid", "wa-key", ""},
		{"GetLIDByPhoneNumber", func(c *Client) error { _, err := c.WhatsApp.Contacts.GetLIDByPhoneNumber(ctx, "3@c.us"); return err },
			"GET", "/whatsapp/v2/wa-1/contacts/get-lid-by-phone-number?phoneNumber=3%40c.us", "wa-key", ""},
		// SMS
		{"SendMessage", func(c *Client) error {
			_, err := c.SMS.SendMessage(ctx, "+201234567890", "Hi", &SendSMSOptions{MaxSendAttempts: 3})
			return err
		}, "POST", "/sms/v1/sms-1/send-message", "sms-key", `{"to":"+201234567890","content":"Hi","max_send_attempts":3}`},
		{"ListMessages", func(c *Client) error {
			_, err := c.SMS.ListMessages(ctx, &ListOptions{Filters: map[string]string{"status": "delivered"}, PerPage: 20})
			return err
		}, "GET", "/sms/v1/sms-1/messages?filter%5Bstatus%5D=delivered&per_page=20", "sms-key", ""},
		{"SMS DeleteMessage", func(c *Client) error { _, err := c.SMS.DeleteMessage(ctx, "m1"); return err },
			"DELETE", "/sms/v1/sms-1/delete-message?messageId=m1", "sms-key", ""},
		{"ListMessageThreads", func(c *Client) error { _, err := c.SMS.ListMessageThreads(ctx, nil); return err },
			"GET", "/sms/v1/sms-1/message-threads", "sms-key", ""},
		{"EditMessageThread", func(c *Client) error { _, err := c.SMS.EditMessageThread(ctx, "t1", false); return err },
			"PUT", "/sms/v1/sms-1/edit-message-thread", "sms-key", `{"messageThreadId":"t1","is_archived":false}`},
		{"DeleteMessageThread", func(c *Client) error { _, err := c.SMS.DeleteMessageThread(ctx, "t1"); return err },
			"DELETE", "/sms/v1/sms-1/delete-message-thread?messageThreadId=t1", "sms-key", ""},
		{"GetHeartbeats", func(c *Client) error {
			_, err := c.SMS.GetHeartbeats(ctx, &ListOptions{Sort: "-created_at"})
			return err
		},
			"GET", "/sms/v1/sms-1/heartbeats?sort=-created_at", "sms-key", ""},
		// OTP
		{"RequestCode", func(c *Client) error {
			_, err := c.OTP.RequestCode(ctx, "1@c.us", &RequestCodeOptions{Length: 6, UseLetter: Bool(false), Name: "Shop"})
			return err
		}, "POST", "/otp/v2/otp-1/request-code", "otp-key", `{"chatId":"1@c.us","length":6,"useLetter":false,"name":"Shop"}`},
		{"ValidateCode", func(c *Client) error { _, err := c.OTP.ValidateCode(ctx, "1@c.us", "123456"); return err },
			"POST", "/otp/v2/otp-1/validate-code", "otp-key", `{"chatId":"1@c.us","code":"123456"}`},
		{"GenerateLink", func(c *Client) error {
			_, err := c.OTP.GenerateLink(ctx, "1@c.us", &GenerateLinkOptions{Expires: 600, Callback: &OTPLinkCallback{Success: "https://x"}})
			return err
		}, "POST", "/otp/v2/otp-1/generate-link", "otp-key", `{"chatId":"1@c.us","expires":600,"callback":{"success":"https://x"}}`},
		// Enterprise
		{"CreateInstance", func(c *Client) error {
			_, err := c.Enterprise.CreateInstance(ctx, "+201234567890", "Support", nil)
			return err
		},
			"POST", "/whatsapp/v2/create-instance", "ent-key", `{"phone":"+201234567890","name":"Support"}`},
		{"GetInstanceHealth", func(c *Client) error { _, err := c.Enterprise.GetInstanceHealth(ctx, "i-1"); return err },
			"GET", "/whatsapp/v2/i-1/health", "ent-key", ""},
		{"RequestNewQRCode", func(c *Client) error { _, err := c.Enterprise.RequestNewQRCode(ctx, "i-1"); return err },
			"POST", "/whatsapp/v2/i-1/request-new-qr-code", "ent-key", ""},
		{"RestartInstance", func(c *Client) error { _, err := c.Enterprise.RestartInstance(ctx, "i-1"); return err },
			"POST", "/whatsapp/v2/i-1/restart", "ent-key", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, got := newTestClient(t, 200, `{"ok":true}`)
			if err := tt.call(c); err != nil {
				t.Fatal(err)
			}
			r := (*got)[0]
			if r.method != tt.method || r.url != tt.url {
				t.Errorf("request = %s %s, want %s %s", r.method, r.url, tt.method, tt.url)
			}
			if r.auth != "Bearer "+tt.auth {
				t.Errorf("Authorization = %q, want Bearer %s", r.auth, tt.auth)
			}
			assertJSON(t, r.body, tt.body)
		})
	}
}

func TestListEndpoints(t *testing.T) {
	ctx := context.Background()
	c, got := newTestClient(t, 200, `[{"lid":"1@lid","pn":"3@c.us"}]`)

	lids, err := c.WhatsApp.Contacts.GetAllKnownLIDs(ctx, &GetAllKnownLIDsOptions{Limit: 10})
	if err != nil || len(lids) != 1 || lids[0]["pn"] != "3@c.us" {
		t.Fatalf("GetAllKnownLIDs = %v, %v", lids, err)
	}
	if _, err := c.Enterprise.ListInstances(ctx, &ListInstancesOptions{Status: "WORKING"}); err != nil {
		t.Fatal(err)
	}
	want := []string{
		"/whatsapp/v2/wa-1/contacts/get-all-known-lids?limit=10",
		"/whatsapp/v2/instances?filter%5Bstatus%5D=WORKING",
	}
	for i, w := range want {
		if (*got)[i].url != w {
			t.Errorf("url = %s, want %s", (*got)[i].url, w)
		}
	}
	if (*got)[1].auth != "Bearer ent-key" {
		t.Errorf("ListInstances used %q", (*got)[1].auth)
	}
}

func TestResultIsDecoded(t *testing.T) {
	c, _ := newTestClient(t, 200, `{"numberExists":true,"chatId":"201234567890@c.us"}`)
	res, err := c.WhatsApp.Contacts.CheckContactExists(context.Background(), "201234567890")
	if err != nil {
		t.Fatal(err)
	}
	if res["numberExists"] != true || res["chatId"] != "201234567890@c.us" {
		t.Errorf("result = %v", res)
	}
}

func TestUploadIsMultipart(t *testing.T) {
	c, got := newTestClient(t, 200, `{}`)
	_, err := c.WhatsApp.SendImage(context.Background(), "1@c.us", MediaFile("logo.png", strings.NewReader("PNGDATA")), &MediaOptions{Caption: "Logo"})
	if err != nil {
		t.Fatal(err)
	}
	r := (*got)[0]
	if !strings.HasPrefix(r.contentType, "multipart/form-data") {
		t.Fatalf("Content-Type = %s", r.contentType)
	}
	for _, want := range []string{`name="chatId"`, "1@c.us", `name="caption"`, "Logo", `name="file"; filename="logo.png"`, "PNGDATA"} {
		if !strings.Contains(string(r.body), want) {
			t.Errorf("multipart body missing %q", want)
		}
	}
}

func TestMultipartBooleansAreOneOrZero(t *testing.T) {
	c, got := newTestClient(t, 200, `{}`)
	_, err := c.WhatsApp.SendLinkCustomPreview(context.Background(), "1@c.us", "Look",
		LinkPreview{Title: "T", Description: "D", URL: "https://x.com", Image: MediaFile("p.png", strings.NewReader("img"))},
		&SendLinkCustomPreviewOptions{HighQuality: true})
	if err != nil {
		t.Fatal(err)
	}
	body := string((*got)[0].body)
	if !strings.Contains(body, "name=\"high_quality\"\r\n\r\n1\r\n") || !strings.Contains(body, `name="preview_image_file"`) {
		t.Errorf("unexpected multipart body:\n%s", body)
	}
}

func TestEmptyMedia(t *testing.T) {
	c, got := newTestClient(t, 200, `{}`)
	if _, err := c.WhatsApp.SendImage(context.Background(), "1@c.us", Media{}, nil); !errors.Is(err, errNeedMedia) {
		t.Errorf("err = %v, want errNeedMedia", err)
	}
	if len(*got) != 0 {
		t.Error("request was sent")
	}
}

func TestAPIError(t *testing.T) {
	c, _ := newTestClient(t, 422, `{"message":"The chat id field is required."}`)
	_, err := c.WhatsApp.SendText(context.Background(), "", "Hi", nil)

	var apiErr *Error
	if !errors.As(err, &apiErr) {
		t.Fatalf("err = %v, want *Error", err)
	}
	if apiErr.StatusCode != 422 || apiErr.Message != "The chat id field is required." {
		t.Errorf("error = %+v", apiErr)
	}
	if err.Error() != "hypersender: 422 The chat id field is required." {
		t.Errorf("Error() = %q", err.Error())
	}
}

func TestAPIErrorWithoutJSON(t *testing.T) {
	c, _ := newTestClient(t, 502, "Bad Gateway")
	_, err := c.WhatsApp.SendText(context.Background(), "1@c.us", "Hi", nil)
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Message != "Bad Gateway" || string(apiErr.Body) != "Bad Gateway" {
		t.Errorf("err = %v", err)
	}
}

func TestEmptyResponse(t *testing.T) {
	c, _ := newTestClient(t, 204, "")
	res, err := c.WhatsApp.StartTyping(context.Background(), "1@c.us")
	if err != nil || res != nil {
		t.Errorf("res, err = %v, %v", res, err)
	}
}

func TestEnvironmentCredentials(t *testing.T) {
	t.Setenv("HYPERSENDER_WHATSAPP_API_KEY", "env-key")
	t.Setenv("HYPERSENDER_WHATSAPP_INSTANCE_ID", "env-1")
	var auth, path string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth, path = r.Header.Get("Authorization"), r.URL.Path
		io.WriteString(w, `{}`)
	}))
	defer srv.Close()

	if _, err := NewClient(WithBaseURL(srv.URL)).WhatsApp.SendText(context.Background(), "1@c.us", "Hi", nil); err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer env-key" || path != "/whatsapp/v2/env-1/send-text" {
		t.Errorf("auth, path = %q, %q", auth, path)
	}
}

func TestMissingCredentials(t *testing.T) {
	t.Setenv("HYPERSENDER_SMS_API_KEY", "")
	t.Setenv("HYPERSENDER_SMS_INSTANCE_ID", "")
	ctx := context.Background()

	_, err := NewClient().SMS.SendMessage(ctx, "+1", "Hi", nil)
	if err == nil || !strings.Contains(err.Error(), "HYPERSENDER_SMS_API_KEY") {
		t.Errorf("err = %v", err)
	}
	_, err = NewClient(WithSMS("k", "")).SMS.SendMessage(ctx, "+1", "Hi", nil)
	if err == nil || !strings.Contains(err.Error(), "HYPERSENDER_SMS_INSTANCE_ID") {
		t.Errorf("err = %v", err)
	}
}

func TestUserAgent(t *testing.T) {
	var ua string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ua = r.Header.Get("User-Agent")
		io.WriteString(w, `{}`)
	}))
	defer srv.Close()
	NewClient(WithBaseURL(srv.URL), WithOTP("k", "i")).OTP.ValidateCode(context.Background(), "1@c.us", "1")
	if ua != "hypersender-go/"+Version {
		t.Errorf("User-Agent = %q", ua)
	}
}

func assertJSON(t *testing.T, got []byte, want string) {
	t.Helper()
	if want == "" {
		if len(got) != 0 {
			t.Errorf("body = %s, want none", got)
		}
		return
	}
	var g, w any
	if err := json.Unmarshal(got, &g); err != nil {
		t.Fatalf("body %q is not JSON: %v", got, err)
	}
	json.Unmarshal([]byte(want), &w)
	if !reflect.DeepEqual(g, w) {
		t.Errorf("body = %s\nwant   %s", got, want)
	}
}
