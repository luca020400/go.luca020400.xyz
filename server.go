package main

import (
	"errors"
	"html/template"
	"io"
	"net/http"
	"server/todo"
	"server/user"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/michaeljs1990/sqlitestore"
)

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type MainData struct {
	Title string
	Data  interface{}
}

func createErrorData(err interface{}) MainData {
	return MainData{
		Title: "Error",
		Data:  err,
	}
}

func renderError(c echo.Context, err error) error {
	return c.Render(http.StatusInternalServerError, "error", createErrorData(err))
}

func createData(title string, data interface{}) MainData {
	return MainData{
		Title: title,
		Data:  data,
	}
}

func renderData(c echo.Context, template string, title string, data interface{}) error {
	return c.Render(http.StatusOK, template, createData(title, data))
}

func Todos(c echo.Context) error {
	return renderData(c, "main", "Todos", nil)
}

func GetTodos(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			renderError(c, err)
		}
	}()

	todos, err := tododb.GetTodos()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "todos", todos)
}

func CreateTodo(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			renderError(c, err)
		}
	}()

	name := c.FormValue("name")
	todo, err := tododb.InsertTodo(name)
	if err != nil {
		return err
	}

	return c.Render(http.StatusCreated, "todo", todo)
}

func DeleteTodo(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			renderError(c, err)
		}
	}()

	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	rows, err := tododb.DeleteTodoByID(idInt)
	if err != nil {
		return err
	}

	if rows == 0 {
		return renderError(c, errors.New("Todo not found"))
	}

	return c.NoContent(http.StatusOK)
}

func CompletedTodo(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			renderError(c, err)
		}
	}()

	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	todo, err := tododb.GetTodoByID(idInt)
	if err != nil {
		return err
	}

	todo.Completed = !todo.Completed
	if _, err = tododb.UpdateTodo(todo); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "todo", todo)
}

var usersdb *user.UserDB
var tododb *todo.TodoDB
var store *sqlitestore.SqliteStore

func main() {
	var err error

	// Setup session store
	if store, err = sqlitestore.NewSqliteStore("sessions.db", "sessions", "/", 3600, []byte("secret")); err != nil {
		panic(err)
	}
	defer store.Close()

	// Setup data DB
	if tododb, err = todo.NewTodoDB(); err != nil {
		panic(err)
	}
	defer tododb.Close()

	// Setup user DB
	if usersdb, err = user.NewUserDB(); err != nil {
		panic(err)
	}
	defer usersdb.Close()

	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	e.Renderer = t

	e.GET("/", Todos)

	e.GET("/api/todos", GetTodos)
	e.POST("/api/todos", CreateTodo)
	e.DELETE("/api/todos/:id", DeleteTodo)
	e.POST("/api/todos/:id/completed", CompletedTodo)

	e.Static("/static", "public/assets")
	e.Logger.Fatal(e.Start(":1323"))
}
