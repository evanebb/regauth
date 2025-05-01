package server

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

			return server.Run(context.Background(), conf)
		},
	}
}

func buildConfiguration(args []string) (*configuration.Configuration, error) {
	if len(args) == 0 {
		return nil, errors.New("no configuration path given")
	}

	configurationFile := args[0]

	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	v.SetEnvPrefix("regauth")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	configuration.SetDefaults(v)
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

	err = conf.IsValid()
	if err != nil {
		return conf, fmt.Errorf("invalid configuration: %w", err)
	}

	return conf, err
}
