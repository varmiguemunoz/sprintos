package auth

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"

	"github.com/markbates/goth"
	githubProvider "github.com/markbates/goth/providers/github"
	googleProvider "github.com/markbates/goth/providers/google"
	"github.com/spf13/viper"
)

type authResult struct {
	user goth.User
	err  error
}

func SetupProviders() {
	port := viper.GetString("auth.callback_port")
	base := fmt.Sprintf("http://localhost:%s/auth", port)

	goth.UseProviders(
		githubProvider.New(
			os.Getenv("GITHUB_CLIENT_ID"),
			os.Getenv("GITHUB_CLIENT_SECRET"),
			base+"/github/callback",
		),
		googleProvider.New(
			os.Getenv("GOOGLE_CLIENT_ID"),
			os.Getenv("GOOGLE_CLIENT_SECRET"),
			base+"/google/callback",
			"email", "profile",
		),
	)
}

func StartLogin(providerName string) (goth.User, error) {
	port := viper.GetString("auth.callback_port")
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
		Addr:    ":" + port,
		Handler: mux,
	}

	go func() {
		_ = server.ListenAndServe()
	}()

	loginURL := fmt.Sprintf("http://localhost:%s/auth/%s", port, providerName)
	openBrowser(loginURL)

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
