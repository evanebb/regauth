package configuration

import (
	"errors"
	"github.com/spf13/viper"
)

type Configuration struct {
	InitialAdmin InitialAdmin
	Log          Log
	HTTP         HTTP
	Database     Database
	Token        Token
	Pat          Pat
}

// SetDefaults sets the defaults for the configuration on a viper.Viper instance.
func SetDefaults(v *viper.Viper) {
	v.SetDefault("log.level", "info")
	v.SetDefault("log.formatter", "text")
	v.SetDefault("http.addr", ":8000")
	v.SetDefault("database.port", 5432)
	v.SetDefault("pat.prefix", "registry_pat_")
}

func (c Configuration) IsValid() error {
	errs := newErrorCollection()

	c.Database.isValid(errs)
	c.Token.isValid(errs)
	c.Pat.isValid(errs)

	if errs.HasErrors() {
		return errs
	}

	return nil
}

type InitialAdmin struct {
	Username string
	Password string
}

type Log struct {
	Level     string
	Formatter string
}

type HTTP struct {
	Addr        string
	Certificate string
	Key         string
	SessionKey  string
}

type Database struct {
	Host     string
	Name     string
	User     string
	Password string
	Port     int
}

func (c Database) isValid(errs *errorCollection) {
	if c.Host == "" {
		errs.Add(errors.New("missing database.host"))
	}

	if c.Name == "" {
		errs.Add(errors.New("missing database.name"))
	}

	if c.User == "" {
		errs.Add(errors.New("missing database.user"))
	}

	if c.Password == "" {
		errs.Add(errors.New("missing database.password"))
	}
}

type Token struct {
	Issuer      string
	Service     string
	Certificate string
	Key         string
	Alg         string
}

func (c Token) isValid(errs *errorCollection) {
	if c.Issuer == "" {
		errs.Add(errors.New("missing token.issuer"))
	}

	if c.Service == "" {
		errs.Add(errors.New("missing token.service"))
	}

	if c.Certificate == "" {
		errs.Add(errors.New("missing token.certificate"))
	}

	if c.Key == "" {
		errs.Add(errors.New("missing token.key"))
	}

	if c.Alg == "" {
		errs.Add(errors.New("missing token.alg"))
	}
}

type Pat struct {
	Prefix string
}

func (c Pat) isValid(errs *errorCollection) {
	if c.Prefix == "" {
		errs.Add(errors.New("missing pat.prefix"))
	}
}
