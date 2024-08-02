package user

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type UserDB struct {
	db                 *sql.DB
	getUsersStmt       *sql.Stmt
	getUserByIDStmt    *sql.Stmt
	insertUserStmt     *sql.Stmt
	updateUserStmt     *sql.Stmt
	deleteUserByIDStmt *sql.Stmt
}

func NewUserDB() (*UserDB, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}

	if err := createUserTable(db); err != nil {
		return nil, err
	}

	getUsersStmt, err := db.Prepare("SELECT id, nick FROM users")
	if err != nil {
		return nil, err
	}

	getUserByIDStmt, err := db.Prepare("SELECT id, nick FROM users WHERE id = ?")
	if err != nil {
		return nil, err
	}

	insertUserStmt, err := db.Prepare("INSERT INTO users (nick, salt, pass) VALUES (?, ?, ?)")
	if err != nil {
		return nil, err
	}

	updateUserStmt, err := db.Prepare("UPDATE users SET nick = ? WHERE id = ?")
	if err != nil {
		return nil, err
	}

	deleteUserByIDStmt, err := db.Prepare("DELETE FROM users WHERE id = ?")
	if err != nil {
		return nil, err
	}

	return &UserDB{
		db:                 db,
		getUsersStmt:       getUsersStmt,
		getUserByIDStmt:    getUserByIDStmt,
		insertUserStmt:     insertUserStmt,
		updateUserStmt:     updateUserStmt,
		deleteUserByIDStmt: deleteUserByIDStmt,
	}, nil
}

func connect() (*sql.DB, error) {
	return sql.Open("sqlite3", "./user.db")
}

func createUserTable(db *sql.DB) error {
	_, err := db.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, nick TEXT, salt TEXT, pass TEXT)")
	return err
}

func (udb *UserDB) Close() {
	udb.getUsersStmt.Close()
	udb.getUserByIDStmt.Close()
	udb.insertUserStmt.Close()
	udb.updateUserStmt.Close()
	udb.deleteUserByIDStmt.Close()
}

func (udb *UserDB) GetUsers() ([]User, error) {
	rows, err := udb.getUsersStmt.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}
	for rows.Next() {
		var user User
		if err := rows.Scan(&user.ID, &user.Nick); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (udb *UserDB) GetUserByID(id int64) (*User, error) {
	row := udb.getUserByIDStmt.QueryRow(id)

	var user User
	if err := row.Scan(&user.ID, &user.Nick); err != nil {
		return nil, err
	}

	return &user, nil
}

func (udb *UserDB) InsertUser(nick, salt, pass string) (*User, error) {
	res, err := udb.insertUserStmt.Exec(nick, salt, pass)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &User{
		ID:   id,
		Nick: nick,
	}, nil
}

func (udb *UserDB) UpdateUser(user *User) (int64, error) {
	res, err := udb.updateUserStmt.Exec(user.Nick, user.ID)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (udb *UserDB) DeleteUserByID(id int64) (int64, error) {
	res, err := udb.deleteUserByIDStmt.Exec(id)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}
