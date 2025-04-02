package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/evanebb/regauth/oas"
	"github.com/fatih/color"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
	"time"
)

func newTokenCmd(client *oas.Client, credentialStore CredentialStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "token",
		Short: "Manage personal access tokens",
		Long:  "Manage personal access tokens.\nPersonal access tokens allow you to authenticate to the regauth API as well as the container registry itself.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	cmd.AddCommand(newListTokensCommand(client))
	cmd.AddCommand(newGetTokenCommand(client))
	cmd.AddCommand(newCreateTokenCommand(client, credentialStore))
	cmd.AddCommand(newDeleteTokenCommand(client))

	return cmd
}

func newListTokensCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all your personal access tokens",
		Long:  "List all your personal access tokens.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			res, err := client.ListPersonalAccessTokens(ctx)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tDESCRIPTION\tPERMISSION\tEXPIRATION\tCREATED")
			for _, token := range res {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", token.ID, token.Description, token.Permission, token.ExpirationDate, token.CreatedAt)
			}
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newGetTokenCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <token>",
		Short: "Get information about a specific personal access token",
		Long:  "Get information about a specific personal access token.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a personal access token ID")
			}

			id, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid ID given: %w", err)
			}

			token, err := client.GetPersonalAccessToken(ctx, oas.GetPersonalAccessTokenParams{
				ID: id,
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tDESCRIPTION\tPERMISSION\tEXPIRATION\tCREATED")
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", token.ID, token.Description, token.Permission, token.ExpirationDate, token.CreatedAt)
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newCreateTokenCommand(client *oas.Client, credentialStore CredentialStore) *cobra.Command {
	var (
		description       string
		permission        string
		expirationDateStr string
		login             bool
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new personal access token",
		Long:  "Create a new personal access token.",
		RunE: func(cmd *cobra.Command, args []string) error {
			expirationDate, err := time.Parse(time.RFC3339, expirationDateStr)
			if err != nil {
				return fmt.Errorf("invalid expiration date %s: %w", expirationDate, err)
			}

			res, err := client.CreatePersonalAccessToken(context.Background(), &oas.PersonalAccessTokenRequest{
				Description:    description,
				Permission:     oas.PersonalAccessTokenRequestPermission(permission),
				ExpirationDate: expirationDate,
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			if login {
				if err := logInUsingToken(credentialStore, res.Token); err != nil {
					fmt.Println("could not log in using new token: " + err.Error())
				} else {
					fmt.Println("logged in using new token!")
				}
			}

			fmt.Println("new personal access token: " + color.New(color.FgGreen, color.Bold).Sprint(res.Token))
			fmt.Println("make sure to copy this token immediately! it cannot be retrieved afterwards.")
			return nil
		},
	}

	cmd.Flags().StringVar(&description, "description", "", "description of the new personal access token")
	_ = cmd.MarkFlagRequired("description")
	cmd.Flags().StringVar(&permission, "permission", "", "permission of the new personal access token, can be 'readOnly', 'readWrite' or 'readWriteDelete'")
	_ = cmd.MarkFlagRequired("permission")
	cmd.Flags().StringVar(&expirationDateStr, "expirationDate", "", "expiration date of the new personal access token, must be a valid RFC3339 date")
	_ = cmd.MarkFlagRequired("expirationDate")
	cmd.Flags().BoolVar(&login, "login", false, "immediately log in using the newly generated token and replace your current credentials")

	return cmd
}

func logInUsingToken(credentialStore CredentialStore, token string) error {
	host, credentials, err := credentialStore.GetCurrent()
	if err != nil {
		return err
	}

	credentials.Token = token
	credentials.Username, credentials.Password = "", ""

	if err := credentialStore.Save(host, credentials); err != nil {
		return err
	}

	return nil
}

func newDeleteTokenCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <token>",
		Short: "Delete a personal access token",
		Long:  "Delete a personal access token.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a personal access token ID")
			}

			id, err := uuid.Parse(args[0])
			if err != nil {
				return fmt.Errorf("invalid ID given: %w", err)
			}

			err = client.DeletePersonalAccessToken(ctx, oas.DeletePersonalAccessTokenParams{
				ID: id,
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("successfully deleted personal access token")
			return nil
		},
	}

	return cmd
}
