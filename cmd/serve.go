package cmd

import (
	"fmt"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/api"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/webhook"
)

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the SprintOS server (REST API + GitHub webhooks)",
	RunE: func(cmd *cobra.Command, args []string) error {
		port, _ := cmd.Flags().GetString("port")

		githubSvc := app.NewGitHubService(DB)
		taskSvc := app.NewTaskService(DB)
		orgSvc := app.NewOrganizationService(DB)

		githubHandler := webhook.NewGitHubHandler(githubSvc, taskSvc, orgSvc)

		apiServer := api.NewServer(DB)

		mux := http.NewServeMux()

		mux.Handle("/webhooks/github", githubHandler)

		mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintln(w, `{"status":"ok","version":"1.0.0"}`)
		})

		apiServer.RegisterRoutes(mux)

		addr := ":" + port
		fmt.Fprintf(os.Stdout, "SprintOS server running on %s\n\n", addr)
		fmt.Fprintf(os.Stdout, "REST API:        http://localhost:%s/api\n", port)
		fmt.Fprintf(os.Stdout, "GitHub webhook:  http://localhost:%s/webhooks/github\n", port)
		fmt.Fprintf(os.Stdout, "Health check:    http://localhost:%s/api/health\n\n", port)
		fmt.Fprintf(os.Stdout, "Generate an API key: sprintos api-key create --name \"my-key\"\n")

		return http.ListenAndServe(addr, mux)
	},
}

func init() {
	serveCmd.Flags().String("port", "8090", "Port to listen on")
	rootCmd.AddCommand(serveCmd)
}
