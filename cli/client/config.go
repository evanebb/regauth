package client

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"text/tabwriter"
)

func newConfigCmd(credentialStore CredentialStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage and inspect configuration for logged-in regauth hosts",
		Long:  "Manage and inspect configuration for logged-in regauth hosts.\nThis allows you to easily switch between multiple different regauth hosts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	cmd.AddCommand(newUseHostCmd(credentialStore))
	cmd.AddCommand(newGetHostCmd(credentialStore))
	cmd.AddCommand(newGetHostsCmd(credentialStore))

	return cmd
}

func newUseHostCmd(credentialStore CredentialStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use-host",
		Short: "Set the current regauth host to manage",
		Long:  "Set the current regauth host to manage.",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("specify the host to use")
			}

			if err := credentialStore.UseHost(args[0]); err != nil {
				return err
			}

			fmt.Println("successfully set current host")
			return nil
		},
	}

	return cmd
}

func newGetHostCmd(credentialStore CredentialStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-host",
		Short: "Gets the current regauth host",
		Long:  "Gets the current regauth host.",
		RunE: func(cmd *cobra.Command, args []string) error {
			host, _, err := credentialStore.GetCurrent()
			if err != nil {
				return err
			}

			fmt.Println(host)
			return nil
		},
	}

	return cmd
}

func newGetHostsCmd(credentialStore CredentialStore) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get-hosts",
		Short: "Gets all known regauth hosts",
		Long:  "Gets all known regauth hosts.",
		RunE: func(cmd *cobra.Command, args []string) error {
			credentials, err := credentialStore.GetAll()
			if err != nil {
				return err
			}

			current, _, err := credentialStore.GetCurrent()
			if err != nil {
				return err
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "CURRENT\tNAME")
			for host := range credentials {
				currentValue := ""
				if host == current {
					currentValue = "*"
				}

				_, _ = fmt.Fprintf(w, "%s\t%s\n", currentValue, host)
			}
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}
