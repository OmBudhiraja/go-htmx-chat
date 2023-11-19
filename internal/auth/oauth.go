package auth

import (
	"fmt"
	"html/template"
	"net/http"

	"github.com/OmBudhiraja/go-htmx-chat/internal/auth/providers"
	"github.com/OmBudhiraja/go-htmx-chat/internal/db"
	"github.com/OmBudhiraja/go-htmx-chat/internal/utils"
	"github.com/go-chi/chi/v5"
	"golang.org/x/oauth2"
)

func InitOauth(router *chi.Mux) {

	router.With(AttachUserToContext).Get("/auth/signin", func(w http.ResponseWriter, r *http.Request) {

		_, exists := r.Context().Value(UserContextKey).(db.User)

		if exists {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}

		tmpl := template.Must(template.ParseFiles("views/signin.html"))

		providersList := make([]string, 0, len(providers.ProvidersMap))

		for k := range providers.ProvidersMap {
			providersList = append(providersList, k)
		}

		tmpl.Execute(w, map[string]interface{}{
			"Providers": providersList,
		})
	})

	router.Post("/auth/{provider}", func(w http.ResponseWriter, r *http.Request) {
		redirectUrl := r.URL.Query().Get("redirectUrl")

		provider := providers.ProvidersMap[chi.URLParam(r, "provider")]

		if provider == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		if redirectUrl == "" {
			redirectUrl = "/"
		}

		state, stateCookie := GenerateState(redirectUrl)
		http.SetCookie(w, stateCookie)

		var url string

		if provider.PKCESupport {
			pkce, pkceCookie := GeneratePKCECode()
			http.SetCookie(w, pkceCookie)
			url = provider.Config.AuthCodeURL(state, oauth2.AccessTypeOnline, oauth2.SetAuthURLParam("code_challenge", pkce), oauth2.SetAuthURLParam("code_challenge_method", "S256"))
		} else {
			url = provider.Config.AuthCodeURL(state, oauth2.AccessTypeOnline)
		}

		http.Redirect(w, r, url, http.StatusSeeOther)
	})

	router.Get("/auth/{provider}/callback", func(w http.ResponseWriter, r *http.Request) {
		provider := providers.ProvidersMap[chi.URLParam(r, "provider")]

		if provider == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

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

		var token *oauth2.Token
		var err error

		if provider.PKCESupport {
			codeVerifier := ValidatePKCECode(w, r)

			if codeVerifier == "" {
				w.WriteHeader(http.StatusBadRequest)
				return
			}

			token, err = provider.Config.Exchange(r.Context(), code, oauth2.SetAuthURLParam("code_verifier", codeVerifier))

			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Failed to exchange token"))
				return
			}

		} else {
			token, err = provider.Config.Exchange(r.Context(), code)
			if err != nil {
				fmt.Println(err)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Failed to exchange token"))
				return
			}
		}

		client := provider.Config.Client(r.Context(), token)

		profile, err := provider.GetProfile(client, token)

		if err != nil {
			fmt.Println(err)
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Failed to get profile"))
			return
		}

		userByAccount, exists, err := db.GetUserByAccount(provider.Name, profile.ProviderId)

		if err != nil {
			fmt.Println(err)
			redirectToError(w, r)
			return
		}

		if exists {
			createSessionAndRedirect(userByAccount.Id, w, r)
			return
		}

		userByEmail, exists, err := db.GetUserByEmail(profile.Email)

		if err != nil {
			fmt.Println(err)
			redirectToError(w, r)
			return
		}

		if exists {
			createAccount(userByEmail.Id, token, profile.ProviderId, provider.Name, provider.Type)
			createSessionAndRedirect(userByEmail.Id, w, r)
			return
		}

		user, err := db.CreateUser(profile.Name, profile.Email, utils.NewNullString(profile.Image))

		if err != nil {
			fmt.Println(err)
			redirectToError(w, r)
			return
		}

		createAccount(user.Id, token, profile.ProviderId, provider.Name, provider.Type)
		createSessionAndRedirect(user.Id, w, r)

	})
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

func createAccount(id string, token *oauth2.Token, providerAccountId string, provider, authType string) {
	db.CreateAccount(id, utils.NewNullString(token.AccessToken), utils.NewNullString(token.RefreshToken), utils.NewNullInteger(token.Expiry.Unix()), provider, providerAccountId, utils.NewNullString(token.Extra("scope").(string)), utils.NewNullString(token.Extra("id_token").(string)), authType)
}
