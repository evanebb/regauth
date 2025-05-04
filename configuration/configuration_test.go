package configuration

import (
	"testing"
)

func TestConfiguration_IsValid(t *testing.T) {
	t.Parallel()

	t.Run("invalid configuration, missing fields", func(t *testing.T) {
		t.Parallel()

		conf := &Configuration{}

		expectedMsg := "missing database.host, missing database.name, missing database.user, missing database.password, missing token.issuer, missing token.service, missing token.certificate, missing token.key, missing token.alg, missing pat.prefix"
		if err := conf.IsValid(); err == nil || err.Error() != expectedMsg {
			t.Errorf("expected error message to be %q, got %q", expectedMsg, err)
		}
	})

	t.Run("valid configuration", func(t *testing.T) {
		t.Parallel()

		conf := &Configuration{
			Database: Database{
				Host:     "host",
				Name:     "name",
				User:     "user",
				Password: "password",
			},
			Token: Token{
				Issuer:      "issuer",
				Service:     "service",
				Certificate: "certificate",
				Key:         "key",
				Alg:         "alg",
			},
			Pat: Pat{
				Prefix: "prefix",
			},
		}

		if err := conf.IsValid(); err != nil {
			t.Errorf("expected no error, got %q", err)
		}
	})
}
