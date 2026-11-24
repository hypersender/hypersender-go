package hypersender_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/hypersender/hypersender-go"
)

func Example() {
	// Reads HYPERSENDER_WHATSAPP_API_KEY and HYPERSENDER_WHATSAPP_INSTANCE_ID.
	client := hypersender.NewClient()

	res, err := client.WhatsApp.SendText(context.Background(), "201234567890@c.us", "Hello from Go!", nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res["queued_request_uuid"])
}

func ExampleNewClient() {
	client := hypersender.NewClient(
		hypersender.WithWhatsApp("your-api-key", "your-instance-id"),
		hypersender.WithSMS("your-sms-api-key", "your-sms-instance-id"),
	)
	_ = client
}

func ExampleWhatsAppService_SendImage() {
	client := hypersender.NewClient()
	ctx := context.Background()

	// From a URL
	client.WhatsApp.SendImage(ctx, "201234567890@c.us", hypersender.MediaURL("https://example.com/menu.png"),
		&hypersender.MediaOptions{Caption: "Today's menu"})

	// From a local file
	f, err := os.Open("invoice.pdf")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	client.WhatsApp.SendFile(ctx, "201234567890@c.us", hypersender.MediaFile("invoice.pdf", f), nil)
}

func ExampleContactsService_CheckContactExists() {
	client := hypersender.NewClient()
	ctx := context.Background()

	res, err := client.WhatsApp.Contacts.CheckContactExists(ctx, "201234567890")
	if err != nil {
		log.Fatal(err)
	}
	if res["numberExists"] == true {
		client.WhatsApp.SendText(ctx, res["chatId"].(string), "Welcome!", nil)
	}
}

func ExampleSMSService_SendMessage() {
	client := hypersender.NewClient()

	client.SMS.SendMessage(context.Background(), "+201234567890", "Your order has shipped!",
		&hypersender.SendSMSOptions{RequestID: "order-1042"})
}

func ExampleOTPService_RequestCode() {
	client := hypersender.NewClient()
	ctx := context.Background()

	client.OTP.RequestCode(ctx, "201234567890@c.us", &hypersender.RequestCodeOptions{Length: 6, Name: "My Shop"})

	res, err := client.OTP.ValidateCode(ctx, "201234567890@c.us", "123456")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(res["success"])
}

func ExampleError() {
	client := hypersender.NewClient()

	_, err := client.WhatsApp.SendText(context.Background(), "201234567890@c.us", "Hi", nil)
	var apiErr *hypersender.Error
	if errors.As(err, &apiErr) {
		fmt.Println(apiErr.StatusCode, apiErr.Message)
	}
}
