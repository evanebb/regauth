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
		Use: "repository",
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
		Use: "list",
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
		Use: "get <namespace/name>",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a repository name")
			}

			split := strings.Split(args[0], "/")
			if len(split) != 2 {
				return errors.New("invalid repository name given")
			}

			repo, err := client.GetRepository(ctx, oas.GetRepositoryParams{
				Namespace: split[0],
				Name:      split[1],
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
		Use: "create",
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
		Use: "delete <namespace/name>",
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()

			if len(args) != 1 {
				return errors.New("specify a repository name")
			}

			split := strings.Split(args[0], "/")
			if len(split) != 2 {
				return errors.New("invalid repository name given")
			}

			err := client.DeleteRepository(ctx, oas.DeleteRepositoryParams{
				Namespace: split[0],
				Name:      split[1],
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
