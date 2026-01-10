# Contributing to Spotigo

Thank you for your interest in contributing to Spotigo!

## Development Setup

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/spotigo.git
   cd spotigo
   ```
   Replace `YOUR_USERNAME` with your GitHub username.
3. Set up Go environment (Go 1.23 or later)
4. Install dependencies:
   ```bash
   go mod download
   ```
5. Run tests:
   ```bash
   # Run all tests (excludes examples)
   go test ./tests/...
   
   # Or run specific test packages
   go test ./tests/unit/...
   go test ./tests/integration/...
   ```
   
   **Note:** The `examples/` directory contains standalone programs and is excluded from test runs. Examples should be run individually with `go run`.

## Code Style

- Use `gofmt` for formatting (or `goimports` for import management)
- Follow Go naming conventions:
  - Exported functions/types: PascalCase
  - Unexported functions/types: camelCase
  - Constants: UPPER_CASE or PascalCase
- Add GoDoc comments for all exported symbols
- Handle errors explicitly (no silent failures)
- Use `context.Context` for cancellation and timeouts

## Testing

- Write unit tests for new features
- Write integration tests for API endpoints
- Aim for >80% code coverage
- Run tests with race detection:
  ```bash
  go test -race ./...
  ```
- Run tests with coverage:
  ```bash
  go test -cover ./...
  ```

## Pull Request Process

1. Create a feature branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```
2. Make your changes
3. Add tests for new functionality
4. Update documentation (README, GoDoc comments, examples)
5. Ensure all tests pass:
   ```bash
   go test ./tests/...
   go build ./...
   ```
6. Commit your changes using [Conventional Commits](https://www.conventionalcommits.org/):
   ```bash
   git commit -m "feat: add new endpoint"
   ```
7. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
8. Submit a pull request with a clear description

## Commit Messages

Use conventional commit format:

- `feat: add new endpoint` - New feature
- `fix: correct error handling` - Bug fix
- `docs: update README` - Documentation
- `test: add unit tests` - Tests
- `refactor: improve code structure` - Code refactoring
- `chore: update dependencies` - Maintenance

## Documentation Requirements

- Add GoDoc comments for all exported functions, types, and methods
- Update README.md if adding new features
- Update examples if API changes
- Ensure all code examples use the correct module path: `github.com/sv4u/spotigo`

## Code Review

- All PRs require review before merging
- Address review comments promptly
- Keep PRs focused and reasonably sized
- Respond to feedback constructively

## Questions?

If you have questions, please:
- Open an issue for discussion
- Check existing issues and PRs
- Review the documentation

Thank you for contributing!
