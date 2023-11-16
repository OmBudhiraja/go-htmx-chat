package providers

import (
	"net/http"

	"golang.org/x/oauth2"
)

type Profile struct {
	ProviderId string
	Name       string
	Email      string
	Image      string
}

type OAuthProvider struct {
	Name        string
	Type        string
	Config      *oauth2.Config
	PKCESupport bool
	GetProfile  func(httpClient *http.Client, token *oauth2.Token) (*Profile, error)
}

var ProvidersMap = map[string]*OAuthProvider{
	"github": &GithubProvider,
	"google": &GoogleProvider,
}
