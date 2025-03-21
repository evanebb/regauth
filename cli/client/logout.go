package client

import (
	"fmt"
	"github.com/spf13/cobra"
)

func newLogoutCmd(credentialStore CredentialStore) *cobra.Command {
	cmd := &cobra.Command{
		Use: "logout",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := credentialStore.Save(Credentials{}); err != nil {
				return err
			}

			fmt.Println("successfully logged out")
			return nil
		},
	}

	return cmd
}
