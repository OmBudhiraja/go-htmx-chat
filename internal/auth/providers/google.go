package providers

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/idtoken"
)

var GoogleProvider OAuthProvider = OAuthProvider{
	Name: "google",
	Type: "oidc",
	Config: &oauth2.Config{
		ClientID:     os.Getenv("GOOGLE_CLIENT_ID"),
		ClientSecret: os.Getenv("GOOGLE_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:5000/auth/google/callback",
		Scopes:       []string{"openid", "profile", "email"},
		Endpoint:     google.Endpoint,
	},
	PKCESupport: true,
	GetProfile: func(httpClient *http.Client, token *oauth2.Token) (*Profile, error) {

		id_token, ok := token.Extra("id_token").(string)

		if !ok || id_token == "" {
			return nil, fmt.Errorf("no id_token field in oauth2 token")
		}

		fmt.Println("ID Token", id_token)

		validatedIdToken, err := idtoken.Validate(context.Background(), id_token, os.Getenv("GOOGLE_CLIENT_ID"))

		if err != nil {
			return nil, err
		}

		return &Profile{
			ProviderId: validatedIdToken.Subject,
			Name:       validatedIdToken.Claims["name"].(string),
			Email:      validatedIdToken.Claims["email"].(string),
			Image:      validatedIdToken.Claims["picture"].(string),
		}, nil
	},
}
