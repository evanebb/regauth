package configuration

import (
	"errors"
	"github.com/spf13/viper"
)

type Configuration struct {
	Log      Log
	HTTP     HTTP
	Database Database
	Token    Token
	Auth     Auth
	Registry Registry
}

// SetDefaults sets the defaults for the configuration on a viper.Viper instance.
func SetDefaults(v *viper.Viper) {
	v.SetDefault("log.level", "info")
	v.SetDefault("log.formatter", "text")
	v.SetDefault("http.addr", ":8000")
	v.SetDefault("database.port", 5432)
	v.SetDefault("auth.local.enabled", true)
}

func (c Configuration) IsValid() error {
	errs := newErrorCollection()

	c.HTTP.isValid(errs)
	c.Database.isValid(errs)
	c.Token.isValid(errs)
	c.Registry.isValid(errs)

	if errs.HasErrors() {
		return errs
	}

	return nil
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

func (c HTTP) isValid(errs *errorCollection) {
	if c.SessionKey == "" {
		errs.Add(errors.New("missing http.sessionKey"))
	}
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

type Auth struct {
	Local LocalAuth
}

type LocalAuth struct {
	enabled bool
}

type Registry struct {
	Host string
}

func (c Registry) isValid(errs *errorCollection) {
	if c.Host == "" {
		errs.Add(errors.New("missing registry.host"))
	}
}
