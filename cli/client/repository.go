package client

import (
	"context"
	"errors"
	"fmt"
	"github.com/evanebb/regauth/oas"
	"github.com/spf13/cobra"
	"os"
	"strings"
	"text/tabwriter"
)

func newRepositoryCmd(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "repository",
		Short: "Manage container registry repositories",
		Long:  "Manage container registry repositories.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Usage()
		},
	}

	cmd.AddCommand(newListRepositoriesCommand(client))
	cmd.AddCommand(newGetRepositoryCommand(client))
	cmd.AddCommand(newCreateRepositoryCommand(client))
	cmd.AddCommand(newDeleteRepositoryCommand(client))

	return cmd
}

func newListRepositoriesCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all your repositories",
		Long:  "List all your repositories.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			res, err := client.ListRepositories(ctx)
			if err != nil {
				fmt.Println(err)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAMESPACE\tNAME\tVISIBILITY\tID")
			for _, repo := range res {
				_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", repo.Namespace, repo.Name, string(repo.Visibility), repo.ID.String())
			}
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newGetRepositoryCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <namespace>/<name>",
		Short: "Get information about a specific repository",
		Long:  "Get information about specific repository.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			namespace, name, err := parseRepositoryNameFromArgs(args)
			if err != nil {
				return err
			}

			repo, err := client.GetRepository(ctx, oas.GetRepositoryParams{
				Namespace: namespace,
				Name:      name,
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			w := tabwriter.NewWriter(os.Stdout, 10, 1, 5, ' ', 0)
			_, _ = fmt.Fprintln(w, "NAMESPACE\tNAME\tVISIBILITY\tID")
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", repo.Namespace, repo.Name, string(repo.Visibility), repo.ID.String())
			_ = w.Flush()

			return nil
		},
	}

	return cmd
}

func newCreateRepositoryCommand(client *oas.Client) *cobra.Command {
	var (
		namespace  string
		name       string
		visibility string
	)

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new repository",
		Long:  "Create a new repository.",
		RunE: func(cmd *cobra.Command, args []string) error {
			repo, err := client.CreateRepository(context.Background(), &oas.RepositoryRequest{
				Namespace:  namespace,
				Name:       name,
				Visibility: oas.RepositoryRequestVisibility(visibility),
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("successfully created repository " + repo.Namespace + "/" + repo.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&namespace, "namespace", "", "namespace of the new repository")
	_ = cmd.MarkFlagRequired("namespace")
	cmd.Flags().StringVar(&name, "name", "", "name of the new repository")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().StringVar(&visibility, "visibility", "", "visibility of the new repository, can be either 'private' or 'public'")
	_ = cmd.MarkFlagRequired("visibility")

	return cmd
}

func newDeleteRepositoryCommand(client *oas.Client) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <namespace>/<name>",
		Short: "Delete a repository",
		Long:  "Delete a repository.",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			namespace, name, err := parseRepositoryNameFromArgs(args)
			if err != nil {
				return err
			}

			err = client.DeleteRepository(ctx, oas.DeleteRepositoryParams{
				Namespace: namespace,
				Name:      name,
			})
			if err != nil {
				fmt.Println(err)
				return nil
			}

			fmt.Println("successfully deleted repository")
			return nil
		},
	}

	return cmd
}

func parseRepositoryNameFromArgs(args []string) (namespace string, name string, err error) {
	argCount := len(args)

	if argCount == 1 {
		// assume repository name is given as <namespace>/<name>
		split := strings.Split(args[0], "/")
		if len(split) != 2 {
			return "", "", errors.New("invalid repository name given")
		}
		return split[0], split[1], nil
	}

	if argCount == 2 {
		// assume repository name is given as <namespace> <name>
		return args[0], args[1], nil
	}

	return "", "", errors.New("specify a repository name")
}
