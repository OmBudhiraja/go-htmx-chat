package db

import (
	"database/sql"
)

type User struct {
	Id    string
	Name  string
	Email string
	Image sql.NullString
}

func GetUserByEmail(email string) (User, bool, error) {
	var user User

	row := DB.QueryRow("SELECT id, name, email, image FROM users WHERE email = $1", email)
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Image)

	if err != nil {
		if err == sql.ErrNoRows {
			return user, false, nil
		}
		return user, false, err
	}

	return user, true, nil
}

func GetUserByAccount(provider, providerAccountId string) (User, bool, error) {
	var user User

	row := DB.QueryRow("SELECT users.id, users.name, users.email, users.image FROM accounts INNER JOIN users ON accounts.user_id = users.id WHERE accounts.provider = $1 AND accounts.provider_account_id = $2", provider, providerAccountId)
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Image)

	if err != nil {
		if err == sql.ErrNoRows {
			return user, false, nil
		}
		return user, false, err
	}

	return user, true, nil
}

func GetUserAndSession(token string) (User, Session, bool, error) {
	var user User
	var session Session

	row := DB.QueryRow("SELECT users.id, users.name, users.email, users.image, sessions.token, sessions.user_id, sessions.expires FROM sessions INNER JOIN users on sessions.user_id = users.id WHERE sessions.token = $1", token)
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Image, &session.Token, &session.UserId, &session.Expires)

	if err != nil {
		if err == sql.ErrNoRows {
			return user, session, false, nil
		}
		return user, session, false, err
	}

	return user, session, true, nil
}

func CreateUser(name, email string, image sql.NullString) (User, error) {
	var user User

	row := DB.QueryRow("INSERT INTO users (name, email, image) VALUES ($1, $2, $3) RETURNING id, name, email, image", name, email, image)
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Image)

	return user, err
}
