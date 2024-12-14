package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID         int
	Nickname   string
	Password   string
}

type Profile struct {
	ID         int
	UserID     int
	FirstName  string
	LastName   string
	Bio        string
}

type Item struct {
	ID         int
	CatalogID  int
	Name       string
	Object3D   []byte
	Photo      []byte
}

func CreateUser(db *pgxpool.Pool, nickname string, password string) (int, error) {
	var id int
	query := `INSERT INTO users (nickname, password) VALUES ($1, $2) RETURNING id`
	err := db.QueryRow(context.Background(), query, nickname, password).Scan(&id)
	if err != nil {
		fmt.Println(err)
		return 0, err
	}
	return id, nil
}

func GetUserByID(db *pgxpool.Pool, id int) (*User, error) {
	query := `SELECT id, nickname, password FROM users WHERE id = $1`
	row := db.QueryRow(context.Background(), query, id)

	var user User
	err := row.Scan(&user.ID, &user.Nickname, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func GetUserByNickName(db *pgxpool.Pool, nickname string) (*User, error) {
	query := `SELECT id, nickname, password FROM users WHERE nickname = $1`
	row := db.QueryRow(context.Background(), query, nickname)

	var user User
	err := row.Scan(&user.ID, &user.Nickname, &user.Password)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func DeleteUserByID(db *pgxpool.Pool, id int) (error) {
	query := `DELETE FROM users WHERE id = $1`
	err := db.QueryRow(context.Background(), query).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

func UserExists(db *pgxpool.Pool, user_id int) (bool, error) {
    var exists bool
    query := `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1)`
    err := db.QueryRow(context.Background(), query, user_id).Scan(&exists)
    if err != nil {
        return false, err
    }
    return exists, nil
}

func CreateProfile(db *pgxpool.Pool, user_id int, first_name string, last_name string, bio string) (int, error) {
	exists, err := UserExists(db, user_id)
	if err != nil || !exists {
		return 0, fmt.Errorf("user with id %d does not exist", user_id)
	}
	var id int
	query := `INSERT INTO profiles (user_id, first_name, last_name, bio) VALUES ($1, $2, $3, $4) RETURNING id`
	err = db.QueryRow(context.Background(), query, user_id, first_name, last_name, bio).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetProfileByID(db *pgxpool.Pool, id int) (*Profile, error) {
	query := `SELECT id, user_id, first_name, last_name, bio WHERE id = $1`
	row := db.QueryRow(context.Background(), query, id)

	var profile Profile
	err := row.Scan(&profile.ID, &profile.UserID, &profile.FirstName, &profile.LastName, &profile.Bio)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func DeleteProfileВyID(db *pgxpool.Pool, id int) (error) {
	query := `DELETE FROM profile WHERE id = $1`
	err := db.QueryRow(context.Background(), query).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}

func CreateItem(db *pgxpool.Pool, catalog_id int, name string, object_3d []byte, photo []byte) (int, error) {
	var id int
	query := `INSERT INTO items (catalog_id, name, object_3d, photo) VALUES ($1, $2, $3, $4) RETURNING id`
	err := db.QueryRow(context.Background(), query, catalog_id, name, object_3d, photo).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func GetItemByID(db *pgxpool.Pool, id int) (*Item, error) {
	query := `SELECT id, catalog_id, name, object_3d, photo WHERE id = $1`
	row := db.QueryRow(context.Background(), query, id)

	var item Item
	err := row.Scan(&item.ID, &item.CatalogID, &item.Name, &item.Object3D, &item.Photo)
	if err != nil {
		return nil, err
	}

	return &item, nil
}

func DeleteItemyID(db *pgxpool.Pool, id int) (error) {
	query := `DELETE FROM item WHERE id = $1`
	err := db.QueryRow(context.Background(), query).Scan(&id)
	if err != nil {
		return err
	}
	return nil
}
