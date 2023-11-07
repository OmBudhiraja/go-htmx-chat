package db

import "time"

type Session struct {
	Token   string
	UserId  string
	Expires time.Time
}

const SessionExpiry = time.Hour * 24 * 30
const SessionUpdateAge = time.Hour * 24

func CreateSession(userId string) (Session, error) {
	expiry := time.Now().Add(SessionExpiry)
	var token string

	row := DB.QueryRow("INSERT INTO sessions (user_id, expires) VALUES ($1, $2) RETURNING token", userId, expiry)
	err := row.Scan(&token)

	return Session{
		Token:   token,
		UserId:  userId,
		Expires: expiry,
	}, err
}

func GetSession(token string) (Session, error) {
	var session Session

	row := DB.QueryRow("SELECT * FROM sessions WHERE token = $1", token)
	err := row.Scan(&session.Token, &session.UserId, &session.Expires)

	return session, err
}

func DeleteSession(token string) {
	DB.Exec("DELETE FROM sessions WHERE token = $1", token)
}

func UpdateSessionExpiry(token string) {
	expiry := time.Now().Add(SessionExpiry)

	DB.Exec("UPDATE sessions SET expires = $1 WHERE token = $2", expiry, token)

}
