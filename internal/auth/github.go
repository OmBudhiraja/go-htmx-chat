package auth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/OmBudhiraja/go-htmx-chat/internal/db"
	"github.com/OmBudhiraja/go-htmx-chat/internal/utils"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

type GithubOAuthRes struct {
	ProviderId int    `json:"id"` // Github ID
	Name       string `json:"name"`
	Email      string `json:"email"`
	AvatarURL  string `json:"avatar_url"`
}

func createSessionAndRedirect(userId string, w http.ResponseWriter, r *http.Request) {
	session, err := db.CreateSession(userId)
	if err != nil {
		fmt.Println(err)
		http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
		return
	}
	redirectUrl := r.Context().Value(RedirectUrlContextKey).(string)
	http.SetCookie(w, CreateCookie(SessionCookieName, session.Token, session.Expires))
	http.Redirect(w, r, redirectUrl, http.StatusTemporaryRedirect)
}

func redirectToError(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/error", http.StatusTemporaryRedirect)
}

func createGithubAccount(id string, token *oauth2.Token, providerAccountId string) {
	db.CreateAccount(id, utils.NewNullString(token.AccessToken), utils.NewNullString(token.RefreshToken), utils.NewNullInteger(token.Expiry.Unix()), "github", providerAccountId, utils.NewNullString(token.Extra("scope").(string)), utils.NewNullString(token.Extra("id_token").(string)))
}

func Github(router *chi.Mux) {
	githubClientId := os.Getenv("GITHUB_CLIENT_ID")
	githubClientSecret := os.Getenv("GITHUB_CLIENT_SECRET")

	if githubClientId == "" || githubClientSecret == "" {
		panic("GITHUB_CLIENT_ID or GITHUB_CLIENT_SECRET not found")
	}

	oauthConfig := &oauth2.Config{
		ClientID:     githubClientId,
		ClientSecret: githubClientSecret,
		RedirectURL:  "http://localhost:5000/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	}

	router.Get("/auth/github", func(w http.ResponseWriter, r *http.Request) {
		redirectUrl := r.URL.Query().Get("redirectUrl")

		if redirectUrl == "" {
			redirectUrl = "/"
		}

		state, stateCookie := GenerateState(redirectUrl)
		url := oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
		http.SetCookie(w, stateCookie)
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	})

	router.Get("/auth/github/callback", func(w http.ResponseWriter, r *http.Request) {

		code := r.URL.Query().Get("code")

		if !ValidateState(w, r) {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Invalid state parameter"))
			return
		}

		if code == "" {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing code parameter"))
			return
		}

		token, err := oauthConfig.Exchange(r.Context(), code)
		if err != nil {
			// return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Failed to exchange token: %s", err.Error()))
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Failed to exchange token"))
			return

		}

		client := oauthConfig.Client(r.Context(), token)
		resp, err := client.Get("https://api.github.com/user")
		if err != nil {
			redirectToError(w, r)
			return
		}
		defer resp.Body.Close()

		var githubUserRes GithubOAuthRes
		var providerAccountId string
		err = json.NewDecoder(resp.Body).Decode(&githubUserRes)

		if err != nil {
			fmt.Println(err)
			redirectToError(w, r)
			return
		}

		providerAccountId = strconv.Itoa(githubUserRes.ProviderId)

		userByAccount, exists, err := db.GetUserByAccount("github", providerAccountId)

		if err != nil {
			fmt.Println(err)
			redirectToError(w, r)
			return
		}

		if exists {
			createSessionAndRedirect(userByAccount.Id, w, r)
			return
		}

		userByEmail, exists, err := db.GetUserByEmail(githubUserRes.Email)

		if err != nil {
			fmt.Println(err)
			redirectToError(w, r)
			return
		}

		if exists {
			createGithubAccount(userByEmail.Id, token, providerAccountId)
			createSessionAndRedirect(userByEmail.Id, w, r)
			return
		}

		user, err := db.CreateUser(githubUserRes.Name, githubUserRes.Email, utils.NewNullString(githubUserRes.AvatarURL))

		if err != nil {
			fmt.Println(err)
			redirectToError(w, r)
			return
		}

		createGithubAccount(user.Id, token, providerAccountId)
		createSessionAndRedirect(user.Id, w, r)

	})

	// router.Get("/me", func(w http.ResponseWriter, r *http.Request) error {
	// 	token := &oauth2.Token{
	// 		AccessToken:  accessToken,
	// 		RefreshToken: refreshToken,
	// 		Expiry:       expiry,
	// 	}

	// 	client := oauthConfig.Client(c.Context(), token)
	// 	resp, err := client.Get("https://api.github.com/user")
	// 	if err != nil {
	// 		// return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Failed to fetch user: %s", err.Error()))
	// 	}
	// 	defer resp.Body.Close()

	// 	var user GitHubUser
	// 	err = json.NewDecoder(resp.Body).Decode(&user)
	// 	if err != nil {
	// 		// return c.Status(fiber.StatusInternalServerError).SendString(fmt.Sprintf("Failed to decode user data: %s", err.Error()))
	// 	}

	// 	return c.JSON(user)
	// })

}
