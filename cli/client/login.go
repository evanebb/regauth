package client

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
)

func newLoginCmd(credentialStore CredentialStore) *cobra.Command {
	var (
		token         string
		tokenStdin    bool
		username      string
		password      string
		passwordStdin bool
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
			credentials := HostCredentials{}

			if tokenStdin || passwordStdin {
				b, err := io.ReadAll(os.Stdin)
				if err != nil {
					return err
				}

				str := string(b)

				if tokenStdin {
					token = str
				} else {
					password = str
				}
			}

			if token != "" {
				credentials.Token = token
			} else if username != "" && password != "" {
				credentials.Username, credentials.Password = username, password
			} else {
				return errors.New("no credentials given")
			}

			if err := credentialStore.Save(host, credentials); err != nil {
				return err
			}

			fmt.Println("successfully logged in")
			return nil
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "personal access token to use for authentication")
	cmd.Flags().BoolVar(&tokenStdin, "token-stdin", false, "read personal access token from stdin")
	cmd.Flags().StringVarP(&username, "username", "u", "", "username to use for authentication")
	cmd.Flags().StringVarP(&password, "password", "p", "", "password to use for authentication")
	cmd.Flags().BoolVar(&passwordStdin, "password-stdin", false, "read password from stdin")

	return cmd
}
