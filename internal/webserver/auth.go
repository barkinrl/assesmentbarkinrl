/**
 * This work is licensed under Apache License, Version 2.0 or later.
 * Please read and understand latest version of Licence.
 */
package webserver

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/kazimsarikaya/assesmentbarkinrl/internal/config"
)

type OIDCConfig struct {
	JwksURI          string `json:"jwks_uri"`
	TokenEndpoint    string `json:"token_endpoint"`
	UserInfoEndpoint string `json:"userinfo_endpoint"`
}

// JWK represents a JSON Web Key.
type JWK struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// JWKS represents a JSON Web Key Set.
type JWKS struct {
	Keys []JWK `json:"keys"`
}

func fetchOIDCConfig(ctx context.Context, wellKnownURL string) (*OIDCConfig, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, wellKnownURL, nil)

	if err != nil {
		slog.Debug("Failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		slog.Debug("Failed to fetch OIDC configuration", "error", err)
		return nil, fmt.Errorf("failed to fetch OIDC configuration: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug("Failed to fetch OIDC configuration", "status", resp.Status)
		return nil, fmt.Errorf("failed to fetch OIDC configuration, status: %s", resp.Status)
	}

	var config OIDCConfig

	if err := json.NewDecoder(resp.Body).Decode(&config); err != nil {
		slog.Debug("Failed to decode OIDC configuration", "error", err)
		return nil, fmt.Errorf("failed to decode OIDC configuration: %w", err)
	}

	return &config, nil
}

// Fetch the JWKS.
func fetchJWKS(ctx context.Context, jwksURL string) (*JWKS, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, jwksURL, nil)

	if err != nil {
		slog.Debug("Failed to create request", "error", err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Accept", "application/json")

	resp, err := http.DefaultClient.Do(req)

	if err != nil {
		slog.Debug("Failed to fetch JWKS", "error", err)
		return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		slog.Debug("Failed to fetch JWKS", "status", resp.Status)
		return nil, fmt.Errorf("failed to fetch JWKS, status: %s", resp.Status)
	}

	var jwks JWKS

	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		slog.Debug("Failed to decode JWKS", "error", err)
		return nil, fmt.Errorf("failed to decode JWKS: %w", err)
	}

	return &jwks, nil
}

// ConvertJWKToPublicKey converts a JWK to an RSA public key.
func convertJWKToPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode the modulus and exponent
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		slog.Debug("Failed to decode modulus", "error", err)
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		slog.Debug("Failed to decode exponent", "error", err)
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert exponent to integer
	e := 0
	for _, b := range eBytes {
		e = e*256 + int(b)
	}

	// Construct the RSA public key
	publicKey := &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: e,
	}

	return publicKey, nil
}

// KeyFunc returns the signing key.
func KeyFunc(ctx context.Context, jwksURL string) jwt.Keyfunc {
	return func(token *jwt.Token) (interface{}, error) {
		// Get the key ID from the token.
		kid, ok := token.Header["kid"].(string)
		if !ok {
			slog.Debug("Missing kid in token header")
			return nil, errors.New("missing kid in token header")
		}

		// Fetch JWKS.
		jwks, err := fetchJWKS(ctx, jwksURL)
		if err != nil {
			slog.Debug("Failed to fetch JWKS", "error", err)
			return nil, fmt.Errorf("failed to fetch JWKS: %w", err)
		}

		// Find the corresponding key.
		for _, jwk := range jwks.Keys {
			if jwk.Kid == kid {
				// Convert the JWK to an RSA public key.
				return convertJWKToPublicKey(jwk)
			}
		}

		slog.Debug("Key not found for kid", "kid", kid)
		return nil, fmt.Errorf("key not found for kid: %s", kid)
	}
}

func validateToken(tokenString string) (bool, error) {
	ctx := context.Background()

	config := config.GetConfig()

	if config.GetOidcIssuer() == "" {
		slog.Debug("OIDC issuer not set")
		return false, errors.New("OIDC issuer not set")
	}

	if config.GetOidcAudience() == "" {
		slog.Debug("OIDC audience not set")
		return false, errors.New("OIDC audience not set")
	}

	wellKnownURL := config.GetOidcIssuer() + "/.well-known/openid-configuration"

	// Fetch OIDC configuration
	oidcConfig, err := fetchOIDCConfig(ctx, wellKnownURL)
	if err != nil {
		slog.Debug("Failed to fetch OIDC configuration", "error", err)
		return false, fmt.Errorf("failed to fetch OIDC configuration: %w", err)
	}

	// Parse the token with the KeyFunc.
	token, err := jwt.Parse(tokenString, KeyFunc(ctx, oidcConfig.JwksURI))
	if err != nil {
		slog.Debug("Token validation failed", "error", err)
		return false, fmt.Errorf("token validation failed: %w", err)
	}

	// Ensure token is valid
	if !token.Valid {
		slog.Debug("Invalid token")
		return false, errors.New("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		slog.Debug("Failed to parse token claims")
		return false, errors.New("failed to parse token claims")
	}

	// Validate expiration (`exp` claim).
	if exp, ok := claims["exp"].(float64); ok {
		expirationTime := time.Unix(int64(exp), 0)
		if time.Now().After(expirationTime) {
			slog.Debug("Token has expired")
			return false, fmt.Errorf("token has expired")
		}
	} else {
		slog.Debug("Missing or invalid exp claim")
		return false, fmt.Errorf("missing or invalid exp claim")
	}

	// Optional: Validate "nbf" (not before) claim.
	if nbf, ok := claims["nbf"].(float64); ok {
		notBeforeTime := time.Unix(int64(nbf), 0)
		if time.Now().Before(notBeforeTime) {
			slog.Debug("Token is not yet valid")
			return false, fmt.Errorf("token is not yet valid")
		}
	}

	// Optional: Validate "iat" (issued at) claim.
	if iat, ok := claims["iat"].(float64); ok {
		issuedAtTime := time.Unix(int64(iat), 0)
		if time.Now().Before(issuedAtTime) {
			slog.Debug("Token issued in the future")
			return false, fmt.Errorf("token issued in the future")
		}
	}

	// Validate claims
	if claims["iss"] != config.GetOidcIssuer() {
		slog.Debug("Invalid issuer", "issuer", claims["iss"])
		return false, errors.New("invalid issuer")
	}

	validAudience := false
	aud, ok := claims["aud"].([]interface{})

	if ok {
		for _, a := range aud {
			if a == config.GetOidcAudience() {
				validAudience = true
				break
			}
		}
	} else {
		if claims["aud"] == config.GetOidcAudience() {
			validAudience = true
		}
	}

	if !validAudience {
		slog.Debug("Invalid audience", "audience", claims["aud"])
		return false, errors.New("invalid audience")
	}

	// Get username

	username, ok := claims["preferred_username"].(string)

	if !ok {
		slog.Debug("Username not found")
		return false, errors.New("username not found")
	}

	// Check if user is in the "admins" group
	groups, ok := claims["groups"].([]interface{})
	if !ok {
		return false, errors.New("groups claim not found or invalid")
	}

	slog.Debug("Groups", "username", username, "groups", groups)

	return true, nil
}

// end of file
