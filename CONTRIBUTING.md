# Contributing

Thank you for your interest in contributing to the Hypersender Go SDK.

## Reporting issues

Open an issue on GitHub with a clear title, steps to reproduce, expected and actual behavior, and your `go version` and SDK version.

## Pull requests

1. Fork the repository and create a branch.
2. Make your change and add a test in `hypersender_test.go`.
3. Make sure the checks pass:

   ```bash
   gofmt -l .        # prints nothing
   go vet ./...
   go test -race ./...
   ```

4. Open a pull request describing the change.

The SDK has no dependencies outside the standard library; please keep it that way. New methods follow the existing style: required values as arguments, optional ones in an `...Options` struct, and a `context.Context` first.
