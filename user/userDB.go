package user

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type UserDB struct {
	db                 *sql.DB
	getUsersStmt       *sql.Stmt
	getUserByNickStmt  *sql.Stmt
	insertUserStmt     *sql.Stmt
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

	getUserByNickStmt, err := db.Prepare("SELECT id, nick, hash FROM users WHERE nick = ?")
	if err != nil {
		return nil, err
	}

	insertUserStmt, err := db.Prepare("INSERT INTO users (nick, hash) VALUES (?, ?)")
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
		getUserByNickStmt:  getUserByNickStmt,
		insertUserStmt:     insertUserStmt,
		deleteUserByIDStmt: deleteUserByIDStmt,
	}, nil
}

func connect() (*sql.DB, error) {
	return sql.Open("sqlite3", "./user.db")
}

func createUserTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			nick TEXT UNIQUE,
			hash TEXT
		)
	`

	_, err := db.Exec(query)
	return err
}

func (udb *UserDB) Close() {
	udb.getUsersStmt.Close()
	udb.getUserByNickStmt.Close()
	udb.insertUserStmt.Close()
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

func (udb *UserDB) GetUserByNick(nick string) (*User, string, error) {
	row := udb.getUserByNickStmt.QueryRow(nick)

	var user User
	var hash string
	if err := row.Scan(&user.ID, &user.Nick, &hash); err != nil {
		return nil, "", err
	}

	return &user, hash, nil
}

func (udb *UserDB) InsertUser(nick, hash string) (*User, error) {
	res, err := udb.insertUserStmt.Exec(nick, hash)
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

func (udb *UserDB) DeleteUserByID(id int64) (int64, error) {
	res, err := udb.deleteUserByIDStmt.Exec(id)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}
