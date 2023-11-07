package db

type Account struct {
	Id                string
	UserId            string
	AccessToken       string
	RefreshToken      string
	ExpiresAt         int64
	Provider          string
	ProviderAccountId string
	Scope             string
	IdToken           string
}

func CreateAccount(details ...any) error {
	_, err := DB.Exec("INSERT INTO accounts (user_id, access_token, refresh_token, expires_at, provider, provider_account_id , scope, id_token) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)", details...)
	return err
}
