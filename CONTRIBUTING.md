# Contributing to QNVS

Thank you for your interest in contributing to QNVS (Quick Node Version Switcher)! We welcome contributions from the community.

## Getting Started

### Prerequisites

- Go 1.24 or later
- Git

### Setting Up the Development Environment

1. Fork the repository on GitHub
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/qnvs.git
   cd qnvs
   ```
3. Install dependencies:
   ```bash
   go mod download
   ```
4. Build the project:
   ```bash
   go build -o qnvs .
   ```

## Making Changes

### Branch Naming

- `feature/` - New features (e.g., `feature/auto-update`)
- `fix/` - Bug fixes (e.g., `fix/windows-symlink`)
- `docs/` - Documentation updates (e.g., `docs/readme-update`)

### Code Style

- Run `go fmt ./...` before committing
- Run `go vet ./...` to catch common issues
- Keep functions focused and well-documented
- Add comments for complex logic

### Testing Your Changes

```bash
# Build and test locally
go build -o qnvs .
./qnvs setup
./qnvs install 22
./qnvs use 22
./qnvs list
```

### Cross-Platform Testing

QNVS supports Windows, macOS, and Linux. If possible, test your changes on multiple platforms or clearly note which platforms you've tested on in your PR.

## Submitting a Pull Request

1. Create a new branch from `main`:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and commit them:
   ```bash
   git add .
   git commit -m "feat: add your feature description"
   ```

3. Push to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```

4. Open a Pull Request on GitHub

### Commit Message Format

We use conventional commits:

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation changes
- `refactor:` - Code refactoring
- `test:` - Adding tests
- `chore:` - Maintenance tasks

Examples:
- `feat: add support for .nvmrc files`
- `fix: resolve Windows junction creation on restricted systems`
- `docs: update installation instructions`

## Reporting Issues

When reporting bugs, please include:

- Operating system and version
- QNVS version (`qnvs version`)
- Steps to reproduce
- Expected vs actual behavior
- Any error messages

## Feature Requests

We welcome feature suggestions! Please:

- Check existing issues to avoid duplicates
- Clearly describe the use case
- Explain why this would benefit other users

## Code of Conduct

Please read and follow our [Code of Conduct](CODE_OF_CONDUCT.md).

## Questions?

Feel free to open an issue for any questions about contributing.

---

Thank you for helping make QNVS better! 🚀