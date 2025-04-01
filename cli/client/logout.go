package client

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
)

func newLogoutCmd(credentialStore CredentialStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logout <host>",
		Short: "Log out from a regauth host",
		Long:  "Log out from a regauth host.\nThis will remove the credentials for the host from the local configuration file.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify the host to log out from")
			}

			if err := credentialStore.Delete(args[0]); err != nil {
				return err
			}

			fmt.Println("successfully logged out")
			return nil
		},
	}

	return cmd
}
