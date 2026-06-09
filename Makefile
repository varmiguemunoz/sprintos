.PHONY: fmt lint build build-prod run tidy check

fmt:
	gofmt -w .

lint:
	golangci-lint run ./...

build:
	go build -o bin/sprintos .

build-prod:
	go build -ldflags "\
		-X github.com/varmiguemunoz/command_pm_app/internal/config.DatabaseURL=$(DATABASE_URL) \
		-X github.com/varmiguemunoz/command_pm_app/internal/config.GitHubClientID=$(OAUTH_GITHUB_CLIENT_ID) \
		-X github.com/varmiguemunoz/command_pm_app/internal/config.GitHubClientSecret=$(OAUTH_GITHUB_CLIENT_SECRET) \
		-X github.com/varmiguemunoz/command_pm_app/internal/config.SMTPHost=$(SMTP_HOST) \
		-X github.com/varmiguemunoz/command_pm_app/internal/config.SMTPPort=$(SMTP_PORT) \
		-X github.com/varmiguemunoz/command_pm_app/internal/config.SMTPFrom=$(SMTP_FROM) \
		-X github.com/varmiguemunoz/command_pm_app/internal/config.SMTPPassword=$(SMTP_PASSWORD)" \
		-o bin/sprintos .

run:
	go run main.go start

tidy:
	go mod tidy

check: fmt lint build
