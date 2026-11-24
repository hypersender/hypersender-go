package hypersender

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
)

// WhatsAppService sends WhatsApp messages. Use it as client.WhatsApp.
type WhatsAppService struct {
	s *service
	// Contacts looks up and saves WhatsApp contacts and LIDs.
	Contacts *ContactsService
}

// SendTextOptions are the optional fields for SendText and SendTextSafe.
type SendTextOptions struct {
	ReplyTo                string `json:"reply_to,omitempty"`                  // ID of the message to reply to
	LinkPreview            bool   `json:"link_preview,omitempty"`              // show a preview for links
	LinkPreviewHighQuality bool   `json:"link_preview_high_quality,omitempty"` // use a high quality preview image
}

type textBody struct {
	ChatID string `json:"chatId"`
	Text   string `json:"text"`
	*SendTextOptions
}

// SendText sends a text message. chatID looks like "201234567890@c.us".
func (w *WhatsAppService) SendText(ctx context.Context, chatID, text string, opts *SendTextOptions) (Result, error) {
	return w.s.object(ctx, request{method: http.MethodPost, path: "/send-text", body: textBody{chatID, text, opts}})
}

// SendTextSafe sends a text message with human-like pacing and typing to lower the risk of a block.
func (w *WhatsAppService) SendTextSafe(ctx context.Context, chatID, text string, opts *SendTextOptions) (Result, error) {
	return w.s.object(ctx, request{method: http.MethodPost, path: "/send-text-safe", body: textBody{chatID, text, opts}})
}

// LinkPreview is the custom preview for SendLinkCustomPreview.
type LinkPreview struct {
	Title       string
	Description string
	URL         string
	Image       Media // optional: MediaURL or MediaFile
}

// SendLinkCustomPreviewOptions are the optional fields for SendLinkCustomPreview.
type SendLinkCustomPreviewOptions struct {
	ReplyTo     string `json:"reply_to,omitempty"`
	HighQuality bool   `json:"high_quality,omitempty"`
}

// SendLinkCustomPreview sends a link with your own preview title, description, and image.
func (w *WhatsAppService) SendLinkCustomPreview(ctx context.Context, chatID, text string, preview LinkPreview, opts *SendLinkCustomPreviewOptions) (Result, error) {
	body := struct {
		ChatID          string `json:"chatId"`
		Text            string `json:"text"`
		PreviewTitle    string `json:"preview_title"`
		PreviewDesc     string `json:"preview_description"`
		PreviewURL      string `json:"preview_url"`
		PreviewImageURL string `json:"preview_image_url,omitempty"`
		*SendLinkCustomPreviewOptions
	}{chatID, text, preview.Title, preview.Description, preview.URL, preview.Image.url, opts}

	r := request{method: http.MethodPost, path: "/send-link-custom-preview", body: body}
	if preview.Image.reader != nil {
		r = request{method: http.MethodPost, path: r.path, form: body, fileField: "preview_image_file", file: preview.Image}
	}
	return w.s.object(ctx, r)
}

// MediaOptions are the optional fields for SendFile, SendImage, and SendVideo.
type MediaOptions struct {
	Filename string `json:"filename,omitempty"` // file name shown to the recipient
	MimeType string `json:"mimetype,omitempty"`
	Caption  string `json:"caption,omitempty"`
	ReplyTo  string `json:"reply_to,omitempty"`
}

// VoiceOptions are the optional fields for SendVoice.
type VoiceOptions struct {
	Caption string `json:"caption,omitempty"`
	ReplyTo string `json:"reply_to,omitempty"`
}

func (w *WhatsAppService) sendMedia(ctx context.Context, path, chatID string, media Media, opts any) (Result, error) {
	if media.isZero() {
		return nil, errNeedMedia
	}
	body := struct {
		ChatID string `json:"chatId"`
		URL    string `json:"url,omitempty"`
	}{chatID, media.url}
	merged, err := merge(body, opts)
	if err != nil {
		return nil, err
	}
	if media.reader != nil {
		return w.s.object(ctx, request{method: http.MethodPost, path: path, form: merged, fileField: "file", file: media})
	}
	return w.s.object(ctx, request{method: http.MethodPost, path: path, body: merged})
}

// SendFile sends a document from a URL or an upload.
func (w *WhatsAppService) SendFile(ctx context.Context, chatID string, file Media, opts *MediaOptions) (Result, error) {
	return w.sendMedia(ctx, "/send-file", chatID, file, opts)
}

// SendImage sends an image from a URL or an upload.
func (w *WhatsAppService) SendImage(ctx context.Context, chatID string, image Media, opts *MediaOptions) (Result, error) {
	return w.sendMedia(ctx, "/send-image", chatID, image, opts)
}

// SendVideo sends a video from a URL or an upload.
func (w *WhatsAppService) SendVideo(ctx context.Context, chatID string, video Media, opts *MediaOptions) (Result, error) {
	return w.sendMedia(ctx, "/send-video", chatID, video, opts)
}

// SendVoice sends a voice note from a URL or an upload.
func (w *WhatsAppService) SendVoice(ctx context.Context, chatID string, voice Media, opts *VoiceOptions) (Result, error) {
	return w.sendMedia(ctx, "/send-voice", chatID, voice, opts)
}

// SendPollOptions are the optional fields for SendPoll.
type SendPollOptions struct {
	MultipleAnswers bool   // let people pick more than one option
	ReplyTo         string // ID of the message to reply to
}

// SendPoll sends a poll.
func (w *WhatsAppService) SendPoll(ctx context.Context, chatID, name string, options []string, opts *SendPollOptions) (Result, error) {
	if opts == nil {
		opts = &SendPollOptions{}
	}
	type poll struct {
		Name            string   `json:"name"`
		Options         []string `json:"options"`
		MultipleAnswers bool     `json:"multipleAnswers"`
	}
	body := struct {
		ChatID  string `json:"chatId"`
		ReplyTo string `json:"reply_to,omitempty"`
		Poll    poll   `json:"poll"`
	}{chatID, opts.ReplyTo, poll{name, options, opts.MultipleAnswers}}
	return w.s.object(ctx, request{method: http.MethodPost, path: "/send-poll", body: body})
}

// ContactCard is a contact to share with SendContactCard.
type ContactCard struct {
	FullName     string `json:"fullName,omitempty"`
	PhoneNumber  string `json:"phoneNumber,omitempty"`
	Organization string `json:"organization,omitempty"`
	WhatsAppID   string `json:"whatsappId,omitempty"`
	VCard        string `json:"vcard,omitempty"` // a raw vCard instead of the fields above
}

// SendContactCard sends one or more contact cards.
func (w *WhatsAppService) SendContactCard(ctx context.Context, chatID string, contacts []ContactCard) (Result, error) {
	body := struct {
		ChatID   string        `json:"chatId"`
		Contacts []ContactCard `json:"contacts"`
	}{chatID, contacts}
	return w.s.object(ctx, request{method: http.MethodPost, path: "/send-contact-card", body: body})
}

// SendLocationOptions are the optional fields for SendLocation.
type SendLocationOptions struct {
	Title   string `json:"title,omitempty"`
	ReplyTo string `json:"reply_to,omitempty"`
}

// SendLocation sends a location pin.
func (w *WhatsAppService) SendLocation(ctx context.Context, chatID string, latitude, longitude float64, opts *SendLocationOptions) (Result, error) {
	body := struct {
		ChatID    string  `json:"chatId"`
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		*SendLocationOptions
	}{chatID, latitude, longitude, opts}
	return w.s.object(ctx, request{method: http.MethodPost, path: "/send-location", body: body})
}

type chatMessage struct {
	ChatID    string `json:"chatId"`
	MessageID string `json:"messageId"`
}

// ForwardMessage forwards a message to a chat.
func (w *WhatsAppService) ForwardMessage(ctx context.Context, chatID, messageID string) (Result, error) {
	return w.s.object(ctx, request{method: http.MethodPost, path: "/forward-message", body: chatMessage{chatID, messageID}})
}

// ReactToMessage reacts to a message with an emoji.
func (w *WhatsAppService) ReactToMessage(ctx context.Context, chatID, messageID, reaction string) (Result, error) {
	body := struct {
		chatMessage
		Reaction string `json:"reaction"`
	}{chatMessage{chatID, messageID}, reaction}
	return w.s.object(ctx, request{method: http.MethodPut, path: "/react-to-message", body: body})
}

// StarMessage stars (star=true) or unstars a message.
func (w *WhatsAppService) StarMessage(ctx context.Context, chatID, messageID string, star bool) (Result, error) {
	body := struct {
		chatMessage
		Star bool `json:"star"`
	}{chatMessage{chatID, messageID}, star}
	return w.s.object(ctx, request{method: http.MethodPut, path: "/star-message", body: body})
}

// DeleteMessage deletes a message for everyone.
func (w *WhatsAppService) DeleteMessage(ctx context.Context, chatID, messageID string) (Result, error) {
	q := url.Values{"chatId": {chatID}, "messageId": {messageID}}
	return w.s.object(ctx, request{method: http.MethodDelete, path: "/delete-message", query: q})
}

// GetQueuedRequest returns a request's status by the queued_request_uuid every send method returns.
func (w *WhatsAppService) GetQueuedRequest(ctx context.Context, uuid string) (Result, error) {
	return w.s.object(ctx, request{method: http.MethodGet, path: "/queued-requests/" + url.PathEscape(uuid)})
}

// SendSeenOptions are the optional fields for SendSeen.
type SendSeenOptions struct {
	Participant string `json:"participant,omitempty"` // in groups, the sender of the message
}

// SendSeen marks a message as seen (blue ticks).
func (w *WhatsAppService) SendSeen(ctx context.Context, chatID, messageID string, opts *SendSeenOptions) (Result, error) {
	body := struct {
		chatMessage
		*SendSeenOptions
	}{chatMessage{chatID, messageID}, opts}
	return w.s.object(ctx, request{method: http.MethodPost, path: "/send-seen", body: body})
}

// ReadChatOptions are the optional fields for ReadChat.
type ReadChatOptions struct {
	Messages int // how many recent messages to mark as read
	Days     int // mark messages from the last N days as read
}

// ReadChat marks a chat's recent messages as read.
func (w *WhatsAppService) ReadChat(ctx context.Context, chatID string, opts *ReadChatOptions) (Result, error) {
	q := url.Values{}
	if opts != nil && opts.Messages > 0 {
		q.Set("messages", strconv.Itoa(opts.Messages))
	}
	if opts != nil && opts.Days > 0 {
		q.Set("days", strconv.Itoa(opts.Days))
	}
	return w.s.object(ctx, request{method: http.MethodPost, path: "/read-chat", query: q, body: chatOnly{chatID}})
}

type chatOnly struct {
	ChatID string `json:"chatId"`
}

// StartTyping shows "typing..." in a chat.
func (w *WhatsAppService) StartTyping(ctx context.Context, chatID string) (Result, error) {
	return w.s.object(ctx, request{method: http.MethodPost, path: "/start-typing", body: chatOnly{chatID}})
}

// StopTyping stops showing "typing..." in a chat.
func (w *WhatsAppService) StopTyping(ctx context.Context, chatID string) (Result, error) {
	return w.s.object(ctx, request{method: http.MethodPost, path: "/stop-typing", body: chatOnly{chatID}})
}
