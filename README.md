# ollamon

Terminal monitor for Ollama nodes.

## install (one-line)

curl -fsSL https://raw.githubusercontent.com/hbasria/ollamon/main/scripts/install.sh | sh

Sadece kurulum (çalıştırmadan):

curl -fsSL https://raw.githubusercontent.com/hbasria/ollamon/main/scripts/install.sh | sh -s -- --no-run

Sürüm vererek kurulum:

curl -fsSL https://raw.githubusercontent.com/hbasria/ollamon/main/scripts/install.sh | sh -s -- --version v0.1.0

Supported targets:

- darwin/amd64
- darwin/arm64 (Apple Silicon)
- linux/amd64
- linux/arm64

## run

go mod tidy
go run ./cmd/ollamon
