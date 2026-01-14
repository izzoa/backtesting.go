# Contributing Guidelines

Thank you for your interest in contributing to Backtesting.go!

## Issues

Before reporting an issue:
- Check if a similar issue is already open
- Check if a similar issue was recently closed — your bug might have been fixed

To have your issue dealt with promptly, construct a [minimal working example](https://en.wikipedia.org/wiki/Minimal_working_example) that exposes the issue clearly and reproducibly.

In case of bugs, please submit **full** error messages and stack traces.

Remember that GitHub Issues supports [markdown](https://www.markdownguide.org/cheat-sheet/), so wrap code in triple-backtick [fenced code blocks](https://www.markdownguide.org/extended-syntax/#syntax-highlighting):

~~~markdown
```go
func main() {
    // your code here
}
```
~~~

## Development Setup

To set up a development environment:

1. [Fork the repository](https://help.github.com/articles/fork-a-repo/)

2. Clone your fork:
   ```bash
   git clone git@github.com:YOUR_USERNAME/backtesting.git
   cd backtesting
   ```

3. Verify Go is installed (requires Go 1.21+):
   ```bash
   go version
   ```

4. Install dependencies:
   ```bash
   go mod download
   ```

## Testing

Please write reasonable unit tests for any new or changed functionality.

Before submitting a PR, ensure all tests pass:

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with race detector
go test -race ./...

# Run benchmarks
go test -bench=. ./...
```

Also ensure idiomatic code style by running:

```bash
# Format code
go fmt ./...

# Run linter (install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
golangci-lint run

# Run vet
go vet ./...
```

## Code Style

- Follow [Effective Go](https://go.dev/doc/effective_go) guidelines
- Use `gofmt` to format your code
- Write clear, descriptive comments for exported types and functions
- Keep functions focused and small
- Use meaningful variable names
- Handle errors explicitly

## Pull Requests

Recommended reading: [How to make your code reviewer fall in love with you](https://mtlynch.io/code-review-love/)

### PR Guidelines

1. **Clear commit messages** — Use the imperative mood ("Add feature" not "Added feature")
2. **One feature per PR** — Keep PRs focused and reviewable
3. **Tests required** — Every new feature must include unit tests
4. **Documentation** — Update documentation for API changes
5. **No breaking changes** — Unless discussed and approved beforehand

### Commit Message Format

```
AREA: Short description (50 chars or less)

More detailed explanation if needed. Wrap at 72 characters.
Explain the problem this commit solves.

Fixes #123
```

Areas: `FIX`, `FEAT`, `PERF`, `DOC`, `TEST`, `REFACTOR`, `BUILD`

## Project Structure

```
backtesting/
├── *.go              # Core backtesting engine
├── data/             # OHLCV data structures and basic indicators
├── lib/              # Extended indicators and utilities
├── optimize/         # Parameter optimization
├── plot/             # HTML chart generation
├── stats/            # Trading statistics
├── examples/         # Runnable examples
└── testdata/         # Test data files
```

## Documentation

- All exported types and functions must have godoc comments
- Examples should be included where helpful
- Run `godoc -http=:6060` to preview documentation locally

## Questions?

Feel free to open a discussion or issue if you have questions about contributing.
