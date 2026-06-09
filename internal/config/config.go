package config

import "os"

var (
	DatabaseURL        = ""
	GitHubClientID     = ""
	GitHubClientSecret = ""
	SMTPHost           = ""
	SMTPPort           = ""
	SMTPFrom           = ""
	SMTPPassword       = ""
)

func GetDatabaseURL() string {
	if DatabaseURL != "" {
		return DatabaseURL
	}
	return os.Getenv("DATABASE_URL")
}

func GetGitHubClientID() string {
	if GitHubClientID != "" {
		return GitHubClientID
	}
	return os.Getenv("OAUTH_GITHUB_CLIENT_ID")
}

func GetGitHubClientSecret() string {
	if GitHubClientSecret != "" {
		return GitHubClientSecret
	}
	return os.Getenv("OAUTH_GITHUB_CLIENT_SECRET")
}

func GetSMTPHost() string {
	if SMTPHost != "" {
		return SMTPHost
	}
	return os.Getenv("SMTP_HOST")
}

func GetSMTPPort() string {
	if SMTPPort != "" {
		return SMTPPort
	}
	return os.Getenv("SMTP_PORT")
}

func GetSMTPFrom() string {
	if SMTPFrom != "" {
		return SMTPFrom
	}
	return os.Getenv("SMTP_FROM")
}

func GetSMTPPassword() string {
	if SMTPPassword != "" {
		return SMTPPassword
	}
	return os.Getenv("SMTP_PASSWORD")
}

var (
	EvolutionAPIURL   = ""
	EvolutionAPIToken = ""
)

func GetEvolutionAPIURL() string {
	if EvolutionAPIURL != "" {
		return EvolutionAPIURL
	}
	return os.Getenv("EVOLUTION_API_URL")
}

func GetEvolutionAPIToken() string {
	if EvolutionAPIToken != "" {
		return EvolutionAPIToken
	}
	return os.Getenv("EVOLUTION_API_TOKEN")
}
