package mcp

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"github.com/varmiguemunoz/sprintos/internal/app"
	"github.com/varmiguemunoz/sprintos/internal/domain"
	"github.com/varmiguemunoz/sprintos/internal/infrastructure/email"
)

func registerOrgTools(
	s *server.MCPServer,
	org *domain.Organization,
	currentUser *domain.User,
	teamSvc *app.TeamService,
	invitationSvc *app.InvitationService,
) {
	s.AddTool(
		mcp.NewTool("invite_user",
			mcp.WithDescription("Invite a user to join the organization by email"),
			mcp.WithString("email", mcp.Required(), mcp.Description("Email address of the person to invite")),
			mcp.WithString("role", mcp.Description("Role to assign: 'owner', 'manager', or 'member'. Defaults to 'manager'")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			args := getArgs(req)
			emailAddr := getString(args, "email")
			if emailAddr == "" {
				return mcp.NewToolResultText("Error: email is required"), nil
			}
			role := getString(args, "role")
			if role == "" {
				role = domain.RoleManager
			}
			if role != domain.RoleOwner && role != domain.RoleManager && role != domain.RoleMember {
				return mcp.NewToolResultText("Error: role must be 'owner', 'manager', or 'member'"), nil
			}

			inv, err := invitationSvc.Create(emailAddr, org.ID, role)
			if err != nil {
				return mcp.NewToolResultText("Error creating invitation: " + err.Error()), nil
			}

			if err := email.SendInvitation(emailAddr, org.Name, inv.Token); err != nil {
				return mcp.NewToolResultText(fmt.Sprintf(
					"Invitation created but email failed to send.\nToken: %s\nJoin command: sprintos join --token %s\nError: %s",
					inv.Token, inv.Token, err.Error(),
				)), nil
			}

			return mcp.NewToolResultText(fmt.Sprintf(
				"✓ Invitation sent to %s as %s.\nJoin command: sprintos join --token %s",
				emailAddr, role, inv.Token,
			)), nil
		},
	)

	s.AddTool(
		mcp.NewTool("list_organization_users",
			mcp.WithDescription("List all users in the organization with their roles"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			members, err := teamSvc.ListMembers(org.ID)
			if err != nil {
				return mcp.NewToolResultText("Error: " + err.Error()), nil
			}

			type userEntry struct {
				ID    uint   `json:"id"`
				Name  string `json:"name"`
				Email string `json:"email"`
				Role  string `json:"role"`
			}

			users := []userEntry{
				{
					ID:    currentUser.ID,
					Name:  currentUser.Name,
					Email: currentUser.Email,
					Role:  domain.RoleOwner,
				},
			}

			for _, m := range members {
				if m.UserID == org.OwnerID {
					continue
				}
				users = append(users, userEntry{
					ID:    m.UserID,
					Name:  m.User.Name,
					Email: m.User.Email,
					Role:  m.Role,
				})
			}

			return mcp.NewToolResultText(marshal(users)), nil
		},
	)
}
