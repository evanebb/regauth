package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/evanebb/regauth/oas"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

func newTeamCmd(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "team",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	cmd.AddCommand(newListTeamsCommand(client))
	cmd.AddCommand(newGetTeamCommand(client))
	cmd.AddCommand(newCreateTeamCommand(client))
	cmd.AddCommand(newDeleteTeamCommand(client))

	cmd.AddCommand(newTeamMemberCmd(client))

	return cmd
}

func newListTeamsCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			res, err := client.ListTeams(ctx)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tID")
			for _, team := range res {
				_, _ = fmt.Fprintf(w, "%s\t%s\n", team.Name, team.ID.String())
			}
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newGetTeamCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "get <team>",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a team name")
			}

			team, err := client.GetTeam(ctx, oas.GetTeamParams{
				Name: args[0],
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAME\tID")
			_, _ = fmt.Fprintf(w, "%s\t%s\n", team.Name, team.ID.String())
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newCreateTeamCommand(client *oas.Client) *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use: "create",
		RunE: func(cmd *cobra.Command, args []string) error {
			team, err := client.CreateTeam(context.Background(), &oas.TeamRequest{
				Name: name,
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("successfully created team " + team.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "name of the new team")
	_ = cmd.MarkFlagRequired("name")

	return cmd
}

func newDeleteTeamCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "delete <team>",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a team name")
			}

			err := client.DeleteTeam(ctx, oas.DeleteTeamParams{
				Name: args[0],
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("successfully deleted team")
			return nil
		},
	}

	return cmd
}

func newTeamMemberCmd(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "member",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	cmd.AddCommand(newListTeamMembersCmd(client))
	cmd.AddCommand(newAddTeamMemberCmd(client))
	cmd.AddCommand(newRemoveTeamMemberCmd(client))

	return cmd
}

func newListTeamMembersCmd(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a team name")
			}

			res, err := client.ListTeamMembers(ctx, oas.ListTeamMembersParams{
				Name: args[0],
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "USERNAME\tROLE\tUSER ID")
			for _, member := range res {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", member.Username, member.Role, member.UserId)
			}
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newAddTeamMemberCmd(client *oas.Client) *cobra.Command {
	var (
		username string
		role     string
	)

	cmd := &cobra.Command{
		Use: "add",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a team name")
			}

			res, err := client.AddTeamMember(ctx,
				&oas.TeamMemberRequest{
					Username: username,
					Role:     oas.TeamMemberRequestRole(role),
				},
				oas.AddTeamMemberParams{
					Name: args[0],
				},
			)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("successfully added user " + res.Username + " to team " + args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "username of the user to add")
	_ = cmd.MarkFlagRequired("username")
	cmd.Flags().StringVar(&role, "role", "", "team role of the user, can be either 'admin' or 'user'")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}

func newRemoveTeamMemberCmd(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "remove",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 2 {
				return errors.New("specify a team name and username")
			}

			err := client.RemoveTeamMember(ctx,
				oas.RemoveTeamMemberParams{
					Name:     args[0],
					Username: args[1],
				},
			)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("successfully removed team member")
			return nil
		},
	}

	return cmd
}
