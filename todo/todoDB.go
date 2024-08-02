package todo

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type TodoDB struct {
	db                 *sql.DB
	getTodosStmt       *sql.Stmt
	getTodoByIDStmt    *sql.Stmt
	insertTodoStmt     *sql.Stmt
	updateTodoStmt     *sql.Stmt
	deleteTodoByIDStmt *sql.Stmt
}

func NewTodoDB() (*TodoDB, error) {
	db, err := connect()
	if err != nil {
		return nil, err
	}

	if err := createTodoTable(db); err != nil {
		return nil, err
	}

	getTodosStmt, err := db.Prepare("SELECT user_id, id, name, completed FROM todos WHERE user_id = ?")
	if err != nil {
		return nil, err
	}

	getTodoByIDStmt, err := db.Prepare("SELECT user_id, id, name, completed FROM todos WHERE user_id = ? AND id = ?")
	if err != nil {
		return nil, err
	}

	insertTodoStmt, err := db.Prepare("INSERT INTO todos (user_id, name, completed) VALUES (?, ?, ?)")
	if err != nil {
		return nil, err
	}

	updateTodoStmt, err := db.Prepare("UPDATE todos SET name = ?, completed = ? WHERE user_id = ? AND id = ?")
	if err != nil {
		return nil, err
	}

	deleteTodoByIDStmt, err := db.Prepare("DELETE FROM todos WHERE user_id = ? AND id = ?")
	if err != nil {
		return nil, err
	}

	return &TodoDB{
		db:                 db,
		getTodosStmt:       getTodosStmt,
		getTodoByIDStmt:    getTodoByIDStmt,
		insertTodoStmt:     insertTodoStmt,
		updateTodoStmt:     updateTodoStmt,
		deleteTodoByIDStmt: deleteTodoByIDStmt,
	}, nil
}

func (tdb *TodoDB) Close() {
	tdb.getTodosStmt.Close()
	tdb.getTodoByIDStmt.Close()
	tdb.insertTodoStmt.Close()
	tdb.updateTodoStmt.Close()
	tdb.deleteTodoByIDStmt.Close()

	tdb.db.Close()
}

func (tdb *TodoDB) GetTodos(userId int64) ([]Todo, error) {
	rows, err := tdb.getTodosStmt.Query(userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	todos := []Todo{}
	for rows.Next() {
		var todo Todo
		err := rows.Scan(&todo.UserID, &todo.ID, &todo.Name, &todo.Completed)
		if err != nil {
			return nil, err
		}

		todos = append(todos, todo)
	}

	return todos, nil
}

func (tdb *TodoDB) GetTodoByID(userId, id int64) (*Todo, error) {
	row := tdb.getTodoByIDStmt.QueryRow(userId, id)

	var todo Todo
	if err := row.Scan(&todo.UserID, &todo.ID, &todo.Name, &todo.Completed); err != nil {
		return nil, err
	}

	return &todo, nil
}

func (tdb *TodoDB) InsertTodo(userId int64, name string) (*Todo, error) {
	res, err := tdb.insertTodoStmt.Exec(userId, name, false)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	return &Todo{
		UserID:    userId,
		ID:        id,
		Name:      name,
		Completed: false,
	}, nil
}

func (tdb *TodoDB) UpdateTodo(name string, completed bool, userId, id int64) (int64, error) {
	res, err := tdb.updateTodoStmt.Exec(name, completed, userId, id)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
}

func (tdb *TodoDB) DeleteTodoByID(userId, id int64) (int64, error) {
	res, err := tdb.deleteTodoByIDStmt.Exec(userId, id)
	if err != nil {
		return 0, err
	}

	return res.RowsAffected()
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
			user_id INTEGER,
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT,
			completed BOOLEAN,
			FOREIGN KEY(user_id) REFERENCES users(id)
		)
	`

	_, err := db.Exec(query)
	return err
}
