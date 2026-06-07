// Package config resolves the endpoint and API token from flags and
// environment variables.
package config

import (
	"errors"
	"os"
	"strings"
)

// Names of the supported environment variables.
const (
	EnvEndpoint = "MOCO_ENDPOINT"
	EnvToken    = "MOCO_TOKEN"
)

// Config holds the resolved connection configuration.
type Config struct {
	// Origin is the normalized base without a path, e.g. "https://whatwedo.mocoapp.com".
	Origin string
	// Token is the API key used for token authentication.
	Token string
}

// BaseURL returns the base URL of the MOCO API v1.
func (c Config) BaseURL() string {
	return c.Origin + "/api/v1"
}

// Resolve determines endpoint and token. Flags take precedence over
// environment variables; an empty flag value counts as unset.
func Resolve(flagEndpoint, flagToken string) (Config, error) {
	endpoint := firstNonEmpty(flagEndpoint, os.Getenv(EnvEndpoint))
	token := firstNonEmpty(flagToken, os.Getenv(EnvToken))

	if endpoint == "" {
		return Config{}, errors.New("no endpoint set: provide --endpoint or " + EnvEndpoint + " (e.g. whatwedo.mocoapp.com)")
	}
	if token == "" {
		return Config{}, errors.New("no token set: provide --token or " + EnvToken)
	}

	return Config{Origin: normalizeOrigin(endpoint), Token: token}, nil
}

// normalizeOrigin returns "https://host" for an endpoint. Any scheme and path
// in the input are dropped; MOCO is always served over https.
func normalizeOrigin(endpoint string) string {
	s := strings.TrimSpace(endpoint)
	if i := strings.Index(s, "://"); i >= 0 {
		s = s[i+3:]
	}
	if i := strings.IndexByte(s, '/'); i >= 0 {
		s = s[:i]
	}
	return "https://" + s
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}
	return ""
}
