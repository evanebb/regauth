package client

import (
	"github.com/evanebb/regauth/oas"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

func NewRootCmd() (*cobra.Command, error) {
	cmd := &cobra.Command{
		Use:   "regauth-cli",
		Short: "regauth-cli is a command-line tool to interact with regauth",
		Long:  "A command-line tool to interact with a regauth instance and manage repositories, personal access tokens, and more.\nMore information can be found at https://github.com/evanebb/regauth.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	configFilePath := filepath.Join(homeDir, ".regauth/auth.json")
	credentialStore, err := NewFileCredentialStore(filepath.Join(configFilePath))
	if err != nil {
		return nil, err
	}

	host, credentials, err := credentialStore.GetCurrent()
	if err != nil {
		return nil, err
	}

	// very fun that this direct conversion works, not guaranteed to keep working since these types aren't really meant
	// to be intertwined
	securitySource := SecuritySource(credentials)

	client, err := oas.NewClient(host, securitySource)
	if err != nil {
		return nil, err
	}

	cmd.AddCommand(newConfigCmd(credentialStore))
	cmd.AddCommand(newLoginCmd(credentialStore))
	cmd.AddCommand(newLogoutCmd(credentialStore))
	cmd.AddCommand(newRepositoryCmd(client))
	cmd.AddCommand(newTokenCmd(client, credentialStore))
	cmd.AddCommand(newTeamCmd(client))
	cmd.AddCommand(newUserCmd(client))

	return cmd, nil
}
