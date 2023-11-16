package providers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

var GithubProvider OAuthProvider = OAuthProvider{
	Name: "github",
	Type: "oauth",
	Config: &oauth2.Config{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:5000/auth/github/callback",
		Scopes:       []string{"user:email"},
		Endpoint:     github.Endpoint,
	},
	PKCESupport: false,
	GetProfile: func(httpClient *http.Client, token *oauth2.Token) (*Profile, error) {

		resp, err := httpClient.Get("https://api.github.com/user")
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		var githubUserRes struct {
			Id        int    `json:"id"`
			Name      string `json:"name"`
			Email     string `json:"email"`
			AvatarURL string `json:"avatar_url"`
		}

		err = json.NewDecoder(resp.Body).Decode(&githubUserRes)

		if err != nil {
			fmt.Println(err)
			return nil, err
		}

		if githubUserRes.Email == "" {
			resp, err = httpClient.Get("https://api.github.com/user/emails")
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}

			err = json.NewDecoder(resp.Body).Decode(&emails)

			fmt.Println(emails)

			if err != nil {
				fmt.Println(err)
				return nil, err
			}

			for _, email := range emails {
				if email.Primary {
					githubUserRes.Email = email.Email
					break
				}
			}

			if githubUserRes.Email == "" {
				return nil, fmt.Errorf("no primary email found")
			}
		}

		return &Profile{
			ProviderId: fmt.Sprintf("%d", githubUserRes.Id),
			Name:       githubUserRes.Name,
			Email:      githubUserRes.Email,
			Image:      githubUserRes.AvatarURL,
		}, nil
	},
}
