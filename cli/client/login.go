package client

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

func newLoginCmd(credentialStore CredentialStore) *cobra.Command {
	var (
		token    string
		username string
		password string
	)

	cmd := &cobra.Command{
		Use:   "login",
		Short: "login",
		Long:  "login",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify a host to log in to")
			}

			host := args[0]
			credentials := Credentials{Host: host}

			if token != "" {
				credentials.Token = token
			} else if username != "" && password != "" {
				credentials.Username, credentials.Password = username, password
			} else {
				return errors.New("no credentials given, please specify either the --username and --password options, or the --token option")
			}

			if err := credentialStore.Save(credentials); err != nil {
				return err
			}

			fmt.Println("successfully logged in")
			return nil
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "personal access token to use for authentication")
	cmd.Flags().StringVarP(&username, "username", "u", "", "username to use for authentication")
	cmd.Flags().StringVarP(&password, "password", "p", "", "password to use for authentication")

	return cmd
}
