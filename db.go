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
	return setupDatabase(tdb.db)
}

func (tdb *TodoDB) GetTodos() ([]Todo, error) {
	return getTodos(tdb.db)
}

func (tdb *TodoDB) GetTodoByID(id int) (*Todo, error) {
	return getTodoByID(tdb.db, id)
}

func (tdb *TodoDB) InsertTodo(todo *Todo) (int64, error) {
	return insertTodo(tdb.db, todo)
}

func (tdb *TodoDB) UpdateTodo(todo *Todo) (int64, error) {
	return updateTodo(tdb.db, todo)
}

func (tdb *TodoDB) DeleteTodoByID(id int64) (int64, error) {
	return deleteTodoByID(tdb.db, id)
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
	if err != nil {
		return err
	}

	return nil
}

func getTodos(db *sql.DB) ([]Todo, error) {
	query, err := db.Prepare("SELECT id, name, completed FROM todos")
	if err != nil {
		return nil, err
	}
	defer query.Close()

	rows, err := query.Query()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.ID, &todo.Name, &todo.Completed)
		if err != nil {
			return nil, err
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

func getTodoByID(db *sql.DB, id int) (*Todo, error) {
	query, err := db.Prepare("SELECT id, name, completed FROM todos WHERE id = ?")
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

func insertTodo(db *sql.DB, todo *Todo) (int64, error) {
	query, err := db.Prepare("INSERT INTO todos (name, completed) VALUES (?, ?)")
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

func updateTodo(db *sql.DB, todo *Todo) (int64, error) {
	query, err := db.Prepare("UPDATE todos SET name = ?, completed = ? WHERE id = ?")
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

func deleteTodoByID(db *sql.DB, id int64) (int64, error) {
	query, err := db.Prepare("DELETE FROM todos WHERE id = ?")
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

func setupDatabase(db *sql.DB) error {
	err := createTodoTable(db)
	if err != nil {
		return err
	}

	return nil
}
