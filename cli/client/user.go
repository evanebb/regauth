package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/evanebb/regauth/oas"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"text/tabwriter"
)

func newUserCmd(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "user",
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
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			res, err := client.ListUsers(ctx)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "USERNAME\tROLE\tID")
			for _, user := range res {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", user.Username, user.Role, user.ID)
			}
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newGetUserCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "get <namespace/name>",
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
			_, _ = fmt.Fprintln(w, "USERNAME\tROLE\tID")
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", user.Username, user.Role, user.ID)
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newCreateUserCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "create",
		RunE: func(cmd *cobra.Command, args []string) error {
			user, err := client.CreateUser(context.Background(), &oas.UserRequest{
				Username: viper.GetString("username"),
				Role:     oas.UserRequestRole(viper.GetString("role")),
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("successfully created user " + user.Username)
			return nil
		},
	}

	cmd.PersistentFlags().String("username", "", "username of the new user")
	_ = cmd.MarkPersistentFlagRequired("username")
	cmd.PersistentFlags().String("role", "", "role of the new user, can be either 'admin' or 'user'")
	_ = cmd.MarkPersistentFlagRequired("role")

	_ = viper.BindPFlags(cmd.PersistentFlags())

	return cmd
}

func newDeleteUserCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "delete <namespace/name>",
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
	cmd := &cobra.Command{
		Use: "change-password",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a username")
			}

			err := client.ChangeUserPassword(ctx,
				&oas.UserPasswordChangeRequest{
					Password: viper.GetString("password"),
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

	cmd.PersistentFlags().String("password", "", "new password for the user")
	_ = cmd.MarkPersistentFlagRequired("password")

	_ = viper.BindPFlags(cmd.PersistentFlags())

	return cmd
}
