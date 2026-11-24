<p align="center"><img width="307" height="63" src="https://raw.githubusercontent.com/hypersender/hypersender-go/main/art/hy.svg" alt="Hypersender Logo"></p>

<p align="center">
<a href="https://pkg.go.dev/github.com/hypersender/hypersender-go"><img src="https://pkg.go.dev/badge/github.com/hypersender/hypersender-go.svg" alt="Go Reference"></a>
<a href="https://github.com/hypersender/hypersender-go/releases"><img src="https://img.shields.io/github/v/release/hypersender/hypersender-go" alt="Latest Stable Version"></a>
<a href="https://goreportcard.com/report/github.com/hypersender/hypersender-go"><img src="https://goreportcard.com/badge/github.com/hypersender/hypersender-go" alt="Go Report Card"></a>
<a href="https://github.com/hypersender/hypersender-go/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-brightgreen.svg" alt="License"></a>
</p>

# Introduction

The official Go SDK for the Hypersender WhatsApp API, SMS API, and OTP API. Send WhatsApp messages from Golang, send SMS from your own Android phone, and verify users with one-time codes. No dependencies beyond the standard library.

# Installation

You can install the package with `go get` (Go 1.24+):

```bash
go get github.com/hypersender/hypersender-go
```

```go
import "github.com/hypersender/hypersender-go"

client := hypersender.NewClient() // reads HYPERSENDER_WHATSAPP_API_KEY and HYPERSENDER_WHATSAPP_INSTANCE_ID
res, err := client.WhatsApp.SendText(ctx, "201234567890@c.us", "Hello from Go!", nil)
```

# Official Documentation

You can find the official documentation at [Hypersender Go SDK Documentation](https://docs.hypersender.com/v2/packages-and-sdks/go/installation), and the full API reference on [pkg.go.dev](https://pkg.go.dev/github.com/hypersender/hypersender-go).

## Changelog

Please see [CHANGELOG](CHANGELOG.md) for more information on what has changed recently.

## Contributing

Please see [CONTRIBUTING](CONTRIBUTING.md) for details.

## Security Vulnerabilities

Please review [our security policy](../../security/policy) on how to report security vulnerabilities.

## Credits

-   [Hypersender](https://github.com/hypersender)
-   [All Contributors](../../contributors)

## License

The MIT License (MIT). Please see [License File](LICENSE) for more information.
