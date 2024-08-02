package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type TodoDB struct {
	db *sql.DB
}

func NewTodoDB() (*TodoDB, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}

	return &TodoDB{db}, nil
}

func (tdb *TodoDB) Setup() error {
	err := createTodoTable(tdb.db)
	if err != nil {
		return err
	}

	return nil
}

func (tdb *TodoDB) GetTodos() ([]*Todo, error) {
	query, err := tdb.db.Prepare("SELECT id, name, completed FROM todos")
	if err != nil {
		return nil, err
	}
	defer query.Close()

	rows, err := query.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []*Todo{}
	for rows.Next() {
		todo := &Todo{}
		err := rows.Scan(&todo.ID, &todo.Name, &todo.Completed)
		if err != nil {
			return nil, err
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

func (tdb *TodoDB) GetTodoByID(id int64) (*Todo, error) {
	query, err := tdb.db.Prepare("SELECT id, name, completed FROM todos WHERE id = ?")
	if err != nil {
		return nil, err
	}
	defer query.Close()

	row := query.QueryRow(id)
	var todo Todo
	err = row.Scan(&todo.ID, &todo.Name, &todo.Completed)
	if err != nil {
		return nil, err
	}

	return &todo, nil
}

func (tdb *TodoDB) InsertTodo(todo *Todo) (int64, error) {
	query, err := tdb.db.Prepare("INSERT INTO todos (name, completed) VALUES (?, ?)")
	if err != nil {
		return 0, err
	}
	defer query.Close()

	res, err := query.Exec(todo.Name, todo.Completed)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func (tdb *TodoDB) UpdateTodo(todo *Todo) (int64, error) {
	query, err := tdb.db.Prepare("UPDATE todos SET name = ?, completed = ? WHERE id = ?")
	if err != nil {
		return 0, err
	}
	defer query.Close()

	res, err := query.Exec(todo.Name, todo.Completed, todo.ID)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (tdb *TodoDB) DeleteTodoByID(id int64) (int64, error) {
	query, err := tdb.db.Prepare("DELETE FROM todos WHERE id = ?")
	if err != nil {
		return 0, err
	}
	defer query.Close()

	res, err := query.Exec(id)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (tdb *TodoDB) Close() error {
	return tdb.db.Close()
}

func connect() (*sql.DB, error) {
	db, err := sql.Open("sqlite3", "data.db")
	if err != nil {
		return nil, err
	}

	return db, nil
}

func createTodoTable(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS todos (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			completed BOOLEAN
		)
	`

	_, err := db.Exec(query)
	return err
}
