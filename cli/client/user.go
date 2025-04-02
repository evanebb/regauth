package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/evanebb/regauth/oas"
	"github.com/spf13/cobra"
	"io"
	"os"
	"text/tabwriter"
)

func newUserCmd(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "user",
		Short: "Manage users",
		Long:  "Manage users. Administrator privileges are required for user management.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	cmd.AddCommand(newListUsersCommand(client))
	cmd.AddCommand(newGetUserCommand(client))
	cmd.AddCommand(newCreateUserCommand(client))
	cmd.AddCommand(newDeleteUserCommand(client))
	cmd.AddCommand(newChangeUserPasswordCommand(client))

	return cmd
}

func newListUsersCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all users",
		Long:  "List all users.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			res, err := client.ListUsers(ctx)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "USERNAME\tROLE\tCREATED\tID")
			for _, user := range res {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", user.Username, user.Role, user.CreatedAt, user.ID)
			}
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newGetUserCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <username>",
		Short: "Get information about a specific user",
		Long:  "Get information about specific user.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a username")
			}

			user, err := client.GetUser(ctx, oas.GetUserParams{
				Username: args[0],
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "USERNAME\tROLE\tCREATED\tID")
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", user.Username, user.Role, user.CreatedAt, user.ID)
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newCreateUserCommand(client *oas.Client) *cobra.Command {
	var (
		username string
		role     string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new user",
		Long:  "Create a new user.",
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := client.CreateUser(context.Background(), &oas.UserRequest{
				Username: username,
				Role:     oas.UserRequestRole(role),
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("successfully created user " + user.Username)
			return nil
		},
	}

	cmd.Flags().StringVar(&username, "username", "", "username of the new user")
	_ = cmd.MarkFlagRequired("username")
	cmd.Flags().StringVar(&role, "role", "", "role of the new user, can be either 'admin' or 'user'")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}

func newDeleteUserCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <username>",
		Short: "Delete a user",
		Long:  "Delete a user.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a username")
			}

			err := client.DeleteUser(ctx, oas.DeleteUserParams{
				Username: args[0],
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("successfully deleted user")
			return nil
		},
	}

	return cmd
}

func newChangeUserPasswordCommand(client *oas.Client) *cobra.Command {
	var (
		password      string
		passwordStdin bool
	)

	cmd := &cobra.Command{
		Use:   "change-password",
		Short: "Change the password for a user",
		Long:  "Change the password for a user.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a username")
			}

			if passwordStdin {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				password = string(b)
			}

			if password == "" {
				return errors.New("no new password given")
			}

			err := client.ChangeUserPassword(ctx,
				&oas.UserPasswordChangeRequest{
					Password: password,
				},
				oas.ChangeUserPasswordParams{
					Username: args[0],
				},
			)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("successfully changed password for user")
			return nil
		},
	}

	cmd.Flags().StringVar(&password, "password", "", "new password for the user")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read new password from stdin")

	return cmd
}
