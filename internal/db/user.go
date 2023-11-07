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

func CreateUser(name, email string, image sql.NullString) (User, error) {
	var user User

	row := DB.QueryRow("INSERT INTO users (name, email, image) VALUES ($1, $2, $3) RETURNING id, name, email, image", name, email, image)
	err := row.Scan(&user.Id, &user.Name, &user.Email, &user.Image)

	return user, err
}
