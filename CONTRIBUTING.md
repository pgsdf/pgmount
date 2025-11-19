# Contributing to PGMount

Thank you for your interest in contributing to PGMount! This document provides guidelines for contributing to the project.

## How to Contribute

### Reporting Bugs

1. **Search existing issues** to avoid duplicates
2. **Create a new issue** with:
   - Clear title and description
   - PGSD/FreeBSD/GhostBSD version
   - Go version (`go version`)
   - Steps to reproduce
   - Expected vs actual behavior
   - Relevant logs (use `--verbose` flag)

### Suggesting Features

1. **Check the TODO.md** to see if it's already planned
2. **Open an issue** describing:
   - The feature and its use case
   - How it would work
   - Why it's useful
   - Any implementation ideas

### Contributing Code

1. **Fork the repository**
2. **Create a feature branch**: `git checkout -b feature-name`
3. **Make your changes**
4. **Follow the coding standards** (see below)
5. **Add tests** for new functionality
6. **Update documentation** as needed
7. **Commit with clear messages**
8. **Push to your fork**
9. **Create a pull request**

## Development Setup

### Prerequisites

```bash
# Install dependencies
pkg install go libnotify

# Optional tools
pkg install hs-pandoc  # For man pages
pkg install golangci-lint  # For linting
```

### Building

```bash
# Clone your fork
git clone https://github.com/pgsdf/pgmount.git
cd pgmount

# Build
make

# Run tests
make test

# Run with verbose logging
./pgmount --verbose
```

## Coding Standards

### Go Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html)
- Use `gofmt` for formatting: `make format`
- Run linter: `make lint`
- Maximum line length: 100 characters
- Use meaningful variable names

### Code Organization

```
pgmount/
├── main.go           # Entry point
├── config/           # Configuration management
├── device/           # Device detection
├── daemon/           # Automount daemon
├── notify/           # Notifications
├── tray/             # Tray icon
└── cmd/              # CLI utilities
```

### Error Handling

- Always check errors
- Return errors, don't panic
- Use `fmt.Errorf` for error wrapping
- Log errors with context

```go
// Good
if err != nil {
    return fmt.Errorf("failed to mount %s: %w", device, err)
}

// Bad
if err != nil {
    panic(err)
}
```

### Logging

- Use `log.Printf` for informational messages
- Use `log.Println` for simple messages
- Prefix with component name for clarity

```go
log.Printf("Device manager: scanning for devices")
log.Printf("Mounted %s at %s", dev.Path, mountPoint)
```

### Testing

- Write tests for new functionality
- Place tests in `*_test.go` files
- Use table-driven tests where appropriate
- Aim for >70% code coverage

```go
func TestDeviceDetection(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected bool
    }{
        {"USB drive", "da0", true},
        {"Internal disk", "ada0", false},
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := isRemovable(tt.input)
            if result != tt.expected {
                t.Errorf("got %v, want %v", result, tt.expected)
            }
        })
    }
}
```

### Documentation

- Add comments for exported functions
- Use godoc format
- Update README.md for user-facing changes
- Update man pages for CLI changes

```go
// Mount mounts a device at the specified mount point.
// Returns an error if the device is already mounted or if
// the mount operation fails.
func Mount(dev *Device, mountPoint string) error {
    // ...
}
```

## Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) format:

```
type(scope): description

[optional body]

[optional footer]
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Adding or updating tests
- `chore`: Maintenance tasks

Examples:

```
feat(daemon): add devd socket monitoring

Replaces polling with real-time devd events for faster
device detection.

Closes #42
```

```
fix(notify): handle missing notify-send gracefully

Instead of crashing, log a warning and disable notifications
if notify-send is not available.
```

## Pull Request Process

1. **Update documentation** if needed
2. **Add tests** for new features
3. **Ensure all tests pass**: `make test`
4. **Format code**: `make format`
5. **Update CHANGELOG.md** with your changes
6. **Reference related issues** in PR description
7. **Wait for review** - maintainers will review your PR

### PR Checklist

- [ ] Code follows project style guidelines
- [ ] Tests added/updated and passing
- [ ] Documentation updated
- [ ] CHANGELOG.md updated
- [ ] Commits have clear messages
- [ ] No merge conflicts

## Code Review

### What We Look For

- **Correctness**: Does it work as intended?
- **Testing**: Are there adequate tests?
- **Style**: Does it follow our guidelines?
- **Documentation**: Is it well documented?
- **Simplicity**: Is it the simplest solution?

### Responding to Feedback

- Be patient and respectful
- Ask for clarification if needed
- Make requested changes
- Push updates to your branch
- Don't force-push after review has started

## FreeBSD-Specific Guidelines

### System Integration

- Use FreeBSD native tools (geom, mount, geli, etc.)
- Don't assume Linux-specific paths or tools
- Test on multiple FreeBSD versions when possible
- Consider backward compatibility

### Filesystem Support

When adding new filesystem support:
1. Check if kernel module is required
2. Add mount options to config defaults
3. Update documentation
4. Test thoroughly

### Device Detection

- Use `geom` for disk enumeration
- Use `gpart` for partition info
- Use `camcontrol` for USB detection
- Handle edge cases gracefully

## Release Process

Maintainers handle releases, but here's the process:

1. Update version in `Makefile` and `main.go`
2. Update `CHANGELOG.md` with release notes
3. Create git tag: `git tag v1.x.x`
4. Build binaries: `make clean && make`
5. Create GitHub release
6. Update FreeBSD ports (if applicable)

## Community

### Communication

- **GitHub Issues**: Bug reports and feature requests
- **Pull Requests**: Code contributions
- **Discussions**: General questions and ideas

### Code of Conduct

- Be respectful and professional
- Welcome newcomers
- Focus on constructive feedback
- Assume good intentions

## Getting Help

Need help contributing?

1. Check existing documentation (README, INSTALL, DESIGN)
2. Look at existing code for examples
3. Open an issue with the "question" label
4. Read the TODO.md for project direction

## Recognition

Contributors will be:
- Listed in CONTRIBUTORS.md (to be created)
- Mentioned in release notes
- Credited in commits

Thank you for contributing to PGMount!
