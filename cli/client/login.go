package client

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func newLoginCmd(credentialStore CredentialStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "login",
		Short: "login",
		Long:  "login",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return cmd.Usage()
			}

			host := args[0]
			credentials := Credentials{Host: host}

			token, username, password := viper.GetString("token"), viper.GetString("username"), viper.GetString("password")
			if token != "" {
				credentials.Token = token
			} else if username != "" && password != "" {
				credentials.Username, credentials.Password = username, password
			} else {
				return cmd.Usage()
			}

			if err := credentialStore.Save(credentials); err != nil {
				return err
			}

			fmt.Println("successfully logged in")
			return nil
		},
	}

	cmd.PersistentFlags().StringP("username", "u", "", "username to use for authentication")
	cmd.PersistentFlags().StringP("password", "p", "", "password to use for authentication")
	cmd.PersistentFlags().StringP("token", "t", "", "personal access token to use for authentication")

	_ = viper.BindPFlags(cmd.PersistentFlags())

	return cmd
}
