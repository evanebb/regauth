package auth

import (
	"crypto"
	"crypto/x509"
	"github.com/lestrrat-go/jwx/v2/jwa"
)

type AccessTokenConfiguration struct {
	Issuer          string
	Service         string
	SigningAlg      jwa.SignatureAlgorithm
	SigningKey      crypto.PrivateKey
	VerificationKey crypto.PublicKey
	SigningCert     *x509.Certificate
}
