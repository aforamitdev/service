package auth

import (
	"crypto/rsa"
	"net/http"
	"service2/foundations/web"

	"github.com/golang-jwt/jwt"
	"github.com/pkg/errors"
)

const Key int = 1

const (
	RoleAdmin = "ADMIN"
	RoleUser  = "USER"
)

type Claims struct {
	jwt.StandardClaims
	Roles []string `json:"roles"`
}

func (c Claims) HasRoles(roles ...string) bool {
	for _, has := range c.Roles {
		for _, want := range roles {
			if want == has {
				return true
			}
		}
	}
	return false
}

type Keys map[string]*rsa.PrivateKey

type PublicKeyLookup func(kid string) (*rsa.PublicKey, error)

type Auth struct {
	algorithm string
	keyFunc   func(t *jwt.Token) (interface{}, error)
	parser    *jwt.Parser
	keys      Keys
}

func (a *Auth) AddKey(privatekey *rsa.PrivateKey, kid string) {
	a.keys[kid] = privatekey
}

func (a *Auth) RemoveKey(kid string) {
	delete(a.keys, kid)
}

func New(algorithm string, lookup PublicKeyLookup, keys Keys) (*Auth, error) {

	if jwt.GetSigningMethod(algorithm) == nil {
		return nil, errors.Errorf("unknown algorithm %v", algorithm)
	}

	keyFunc := func(t *jwt.Token) (interface{}, error) {

		kid, ok := t.Header["kid"]

		if !ok {
			return nil, errors.New("missing ket id (kid) in token header")
		}

		kidID, ok := kid.(string)

		if !ok {
			return nil, errors.New("user token key id (kid) must be string")
		}
		return lookup(kidID)
	}

	parser := jwt.Parser{
		ValidMethods: []string{algorithm},
	}

	a := Auth{
		algorithm: algorithm,
		keyFunc:   keyFunc,
		parser:    &parser,
		keys:      keys,
	}
	return &a, nil

}

func (a *Auth) GenerateToken(kid string, claims Claims) (string, error) {

	method := jwt.GetSigningMethod(a.algorithm)

	tkn := jwt.NewWithClaims(method, claims)

	tkn.Header["kid"] = kid

	privatekey, ok := a.keys[kid]

	if !ok {
		return "", errors.New("kid lookup failed")
	}

	str, err := tkn.SignedString(privatekey)
	if err != nil {
		return "", errors.Wrap(err, "signing token")
	}

	return str, nil

}

func (a *Auth) ValidateToken(tokenStr string) (Claims, error) {

	var claims Claims

	token, err := a.parser.ParseWithClaims(tokenStr, &claims, a.keyFunc)

	if err != nil {
		return Claims{}, errors.Wrap(err, "parsing token")
	}

	if !token.Valid {
		return Claims{}, errors.New("invalid token")
	}
	return claims, nil

}

var ErrForbidden = web.NewRequestError(errors.New("you are not authorized for that action"), http.StatusForbidden)
