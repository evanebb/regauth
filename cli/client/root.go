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
		Short: "`regauth-cli`",
		Long:  "`regauth-cli`",
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

	securitySource := SecuritySource{
		Token:    credentials.Token,
		Username: credentials.Username,
		Password: credentials.Password,
	}

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
