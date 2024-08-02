package main

import (
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

func renderError(c echo.Context, err interface{}) error {
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
	todos, err := tododb.GetTodos()
	if err != nil {
		return renderError(c, err)
	}

	return c.Render(http.StatusOK, "todos", todos)
}

func CreateTodo(c echo.Context) error {
	name := c.FormValue("name")
	todo := &Todo{
		Name: name,
	}
	id, err := tododb.InsertTodo(todo)
	if err != nil {
		return renderError(c, err)
	}

	todo.ID = id
	return c.Render(http.StatusCreated, "todo", todo)
}

func DeleteTodo(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return renderError(c, err)
	}

	rows, err := tododb.DeleteTodoByID(idInt)
	if err != nil {
		return renderError(c, err)
	}

	if rows == 0 {
		return renderError(c, "Todo not found")
	}

	return c.NoContent(http.StatusOK)
}

func CompletedTodo(c echo.Context) error {
	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return renderError(c, err)
	}

	todo, err := tododb.GetTodoByID(idInt)
	if err != nil {
		return renderError(c, err)
	}

	todo.Completed = !todo.Completed
	_, err = tododb.UpdateTodo(todo)
	if err != nil {
		return renderError(c, err)
	}

	return c.Render(http.StatusOK, "todo", todo)
}

var tododb *TodoDB

func main() {
	// Setup DB
	var err error
	tododb, err = NewTodoDB()
	if err != nil {
		panic(err)
	}
	defer tododb.Close()
	tododb.Setup()

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
