package cli

import (
	"context"
	"errors"
	"fmt"
	"github.com/evanebb/regauth/configuration"
	"github.com/evanebb/regauth/server"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"strings"
)

func newServeCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "serve <config>",
		Short: "`serve` runs the registry authorization server",
		Long:  "`serve` runs the registry authorization server",
		RunE: func(cmd *cobra.Command, args []string) error {
			conf, err := buildConfiguration(args)
			if err != nil {
				return err
			}

			ctx := context.Background()
			s, err := server.New(ctx, conf)
			if err != nil {
				return err
			}

			return s.ListenAndServe(ctx)
		},
	}
}

func buildConfiguration(args []string) (*configuration.Configuration, error) {
	if len(args) == 0 {
		return nil, errors.New("no configuration path given")
	}

	configurationFile := args[0]

	v := viper.New()
	v.SetEnvPrefix("regauth")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	setDefaults(v)

	v.SetConfigFile(configurationFile)
	err := v.ReadInConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to read configuration file %s: %w", configurationFile, err)
	}

	conf := &configuration.Configuration{}

	err = v.Unmarshal(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal configuration: %w", err)
	}

	return conf, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("log.level", "info")
	v.SetDefault("log.formatter", "text")
	v.SetDefault("addr", ":80")
	v.SetDefault("database.port", 5432)
	v.SetDefault("auth.local.enabled", true)
}
