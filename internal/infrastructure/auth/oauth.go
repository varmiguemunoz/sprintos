package auth

import (
	"context"
	"fmt"
	"net/http"
	"os/exec"
	"runtime"

	"github.com/markbates/goth"
	githubProvider "github.com/markbates/goth/providers/github"
	"github.com/varmiguemunoz/sprintos/internal/config"
)

type authResult struct {
	user goth.User
	err  error
}

func SetupProviders() {
	callbackURL := "http://localhost:8080/auth/github/callback"

	goth.UseProviders(
		githubProvider.New(
			config.GetGitHubClientID(),
			config.GetGitHubClientSecret(),
			callbackURL,
		),
	)
}

func StartLogin(providerName string) (goth.User, error) {
	result := make(chan authResult, 1)

	var gothSession goth.Session

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/"+providerName, func(w http.ResponseWriter, r *http.Request) {
		provider, err := goth.GetProvider(providerName)
		if err != nil {
			result <- authResult{err: fmt.Errorf("provider not found: %w", err)}
			return
		}

		sess, err := provider.BeginAuth("state")
		if err != nil {
			result <- authResult{err: fmt.Errorf("could not begin auth: %w", err)}
			return
		}

		gothSession = sess

		url, err := sess.GetAuthURL()
		if err != nil {
			result <- authResult{err: fmt.Errorf("could not get auth url: %w", err)}
			return
		}

		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	mux.HandleFunc("/auth/"+providerName+"/callback", func(w http.ResponseWriter, r *http.Request) {
		provider, err := goth.GetProvider(providerName)
		if err != nil {
			result <- authResult{err: fmt.Errorf("provider not found: %w", err)}
			return
		}

		_, err = gothSession.Authorize(provider, r.URL.Query())
		if err != nil {
			result <- authResult{err: fmt.Errorf("authorization failed: %w", err)}
			return
		}

		user, err := provider.FetchUser(gothSession)
		if err != nil {
			result <- authResult{err: fmt.Errorf("could not fetch user: %w", err)}
			return
		}

		fmt.Fprintln(w, "Login successful. You can close this tab.")
		result <- authResult{user: user}
	})

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	go func() {
		_ = server.ListenAndServe()
	}()

	openBrowser("http://localhost:8080/auth/" + providerName)

	res := <-result

	_ = server.Shutdown(context.Background())

	if res.err != nil {
		return goth.User{}, res.err
	}

	return res.user, nil
}

func openBrowser(url string) {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("open", url)
	case "linux":
		cmd = exec.Command("xdg-open", url)
	case "windows":
		cmd = exec.Command("cmd", "/c", "start", url)
	}

	if cmd != nil {
		_ = cmd.Start()
	}
}
