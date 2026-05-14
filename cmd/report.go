package cmd

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/auth"
)

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a project health report",
	RunE: func(cmd *cobra.Command, args []string) error {
		projectFlag, _ := cmd.Flags().GetUint("project")
		days, _ := cmd.Flags().GetInt("completed")

		session, err := auth.LoadSession()
		if err != nil {
			return fmt.Errorf("not logged in — run `sprintos start` first")
		}

		userSvc := app.NewUserService(DB)
		orgSvc := app.NewOrganizationService(DB)
		projectSvc := app.NewProjectService(DB)
		taskSvc := app.NewTaskService(DB)
		stateSvc := app.NewStateService(DB)
		teamSvc := app.NewTeamService(DB)

		user, err := userSvc.GetByEmail(session.Email)
		if err != nil {
			return fmt.Errorf("could not load user: %w", err)
		}

		org, err := orgSvc.GetByOwnerID(user.ID)
		if err != nil {
			return fmt.Errorf("could not load organization: %w", err)
		}

		var projectIDs []uint
		if projectFlag > 0 {
			projectIDs = []uint{projectFlag}
		} else {
			projects, _ := projectSvc.ListByOrganization(org.ID)
			for _, p := range projects {
				projectIDs = append(projectIDs, p.ID)
			}
		}

		members, _ := teamSvc.ListMembers(org.ID)
		memberNames := make(map[uint]string)
		for _, m := range members {
			memberNames[m.UserID] = m.User.Name
		}

		fmt.Printf("\nSprintOS Report — %s — %s\n", org.Name, time.Now().Format("Mon Jan 2 2006"))
		fmt.Println(strings.Repeat("─", 55))

		now := time.Now()

		for _, pid := range projectIDs {
			proj, err := projectSvc.GetByID(pid)
			if err != nil {
				continue
			}
			tasks, _ := taskSvc.ListByProject(pid)
			states, _ := stateSvc.ListByProject(pid)

			stateNames := make(map[uint]string)
			stateCount := make(map[string]int)
			isDone := make(map[uint]bool)
			for _, st := range states {
				stateNames[st.ID] = st.Name
				isDone[st.ID] = st.IsDone
			}

			var overdue, unassigned, completed int
			workload := make(map[uint]int)
			var leadTimes []float64
			var cycleSeconds []int64

			for _, t := range tasks {
				stateCount[stateNames[t.StateID]]++
				if t.CompletedAt != nil {
					completed++
					leadTimes = append(leadTimes, t.CompletedAt.Sub(t.CreatedAt).Hours()/24)
				}
				if t.AssignedTo == nil {
					unassigned++
				} else {
					workload[*t.AssignedTo]++
				}
				if t.DueDate != nil && t.DueDate.Before(now) && t.CompletedAt == nil {
					overdue++
				}
			}

			var transitions []struct{ SecondsInFromState int64 }
			DB.Table("state_transitions").
				Joins("JOIN tasks ON tasks.id = state_transitions.task_id").
				Where("tasks.project_id = ? AND state_transitions.deleted_at IS NULL", pid).
				Select("state_transitions.seconds_in_from_state").
				Scan(&transitions)
			for _, tr := range transitions {
				cycleSeconds = append(cycleSeconds, tr.SecondsInFromState)
			}

			fmt.Printf("\n  %s\n", proj.Name)
			fmt.Printf("  Total tasks: %d  •  Completed: %d  •  Overdue: %d  •  Unassigned: %d\n",
				len(tasks), completed, overdue, unassigned)

			fmt.Printf("\n  By state:\n")
			for _, st := range states {
				fmt.Printf("    %-20s  %d\n", st.Name, stateCount[st.Name])
			}

			if len(leadTimes) > 0 {
				avg := 0.0
				for _, l := range leadTimes {
					avg += l
				}
				avg /= float64(len(leadTimes))
				fmt.Printf("\n  Avg lead time: %.1f days (creation → done)\n", avg)
			}

			if len(cycleSeconds) > 0 {
				var total int64
				for _, s := range cycleSeconds {
					total += s
				}
				avgHours := float64(total) / float64(len(cycleSeconds)) / 3600
				fmt.Printf("  Avg cycle time per state: %.1f hours\n", avgHours)
			}

			if len(workload) > 0 {
				fmt.Printf("\n  Team workload:\n")
				type kv struct {
					name  string
					count int
				}
				var sorted []kv
				for uid, count := range workload {
					name := memberNames[uid]
					if name == "" {
						name = fmt.Sprintf("user#%d", uid)
					}
					sorted = append(sorted, kv{name, count})
				}
				sort.Slice(sorted, func(i, j int) bool { return sorted[i].count > sorted[j].count })
				for _, kv := range sorted {
					bar := strings.Repeat("█", kv.count)
					fmt.Printf("    %-20s  %s (%d)\n", kv.name, bar, kv.count)
				}
			}

			if days > 0 {
				since := now.AddDate(0, 0, -days)
				fmt.Printf("\n  Completed in last %d days:\n", days)
				found := 0
				for _, t := range tasks {
					if t.CompletedAt != nil && t.CompletedAt.After(since) {
						lead := t.CompletedAt.Sub(t.CreatedAt).Hours() / 24
						fmt.Printf("    #%d %s  (%.1f days)\n", t.TaskNumber, t.Title, lead)
						found++
					}
				}
				if found == 0 {
					fmt.Printf("    none\n")
				}
			}
		}

		fmt.Println()
		return nil
	},
}

func init() {
	reportCmd.Flags().Uint("project", 0, "Project ID (default: all projects)")
	reportCmd.Flags().Int("completed", 0, "Also show tasks completed in last N days")
	rootCmd.AddCommand(reportCmd)
}
