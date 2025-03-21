package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/evanebb/regauth/oas"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"text/tabwriter"
	"time"
)

func newTokenCmd(client *oas.Client, credentialStore CredentialStore) *cobra.Command {
	cmd := &cobra.Command{
		Use: "token",
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
		Use: "list",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			res, err := client.ListPersonalAccessTokens(ctx)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "ID\tDESCRIPTION\tPERMISSION\tEXPIRATION")
			for _, token := range res {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", token.ID, token.Description, token.Permission, token.ExpirationDate)
			}
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newGetTokenCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "get <token>",
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
			_, _ = fmt.Fprintln(w, "ID\tDESCRIPTION\tPERMISSION\tEXPIRATION")
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", token.ID, token.Description, token.Permission, token.ExpirationDate)
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newCreateTokenCommand(client *oas.Client, credentialStore CredentialStore) *cobra.Command {
	cmd := &cobra.Command{
		Use: "create",
		RunE: func(cmd *cobra.Command, args []string) error {
			expirationDate, err := time.Parse(time.RFC3339, viper.GetString("expirationDate"))
			if err != nil {
				return fmt.Errorf("invalid expiration date %s: %w", viper.GetString("expirationDate"), err)
			}

			res, err := client.CreatePersonalAccessToken(context.Background(), &oas.PersonalAccessTokenRequest{
				Description:    viper.GetString("description"),
				Permission:     oas.PersonalAccessTokenRequestPermission(viper.GetString("permission")),
				ExpirationDate: expirationDate,
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			if viper.GetBool("login") {
				if err := logInUsingToken(credentialStore, res.Token); err != nil {
					fmt.Println("could not log in using new token: " + err.Error())
				} else {
					fmt.Println("logged in using new token")
				}
			}

			fmt.Println("new personal access token: " + res.Token)
			fmt.Println("make sure to copy this token immediately! it cannot be retrieved afterwards")
			return nil
		},
	}

	cmd.PersistentFlags().String("description", "", "description of the new personal access token")
	_ = cmd.MarkPersistentFlagRequired("description")
	cmd.PersistentFlags().String("permission", "", "permission of the new personal access token, can be 'readOnly', 'readWrite' or 'readWriteDelete'")
	_ = cmd.MarkPersistentFlagRequired("permission")
	cmd.PersistentFlags().String("expirationDate", "", "expiration date of the new personal access token, must be a valid RFC3339 date")
	_ = cmd.MarkPersistentFlagRequired("expirationDate")
	cmd.PersistentFlags().Bool("login", false, "immediately log in using the newly generated token and replace your current credentials")

	_ = viper.BindPFlags(cmd.PersistentFlags())

	return cmd
}

func logInUsingToken(credentialStore CredentialStore, token string) error {
	credentials, err := credentialStore.Get()
	if err != nil {
		return err
	}

	credentials.Token = token
	credentials.Username, credentials.Password = "", ""

	if err := credentialStore.Save(credentials); err != nil {
		return err
	}

	return nil
}

func newDeleteTokenCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use: "delete <token>",
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
