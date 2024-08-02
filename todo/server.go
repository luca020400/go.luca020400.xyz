package todo

import (
	"errors"
	"net/http"
	"server/util"
	"strconv"

	"github.com/labstack/echo/v4"
)

type server struct {
	tododb *TodoDB
}

func RegisterHandlers(e *echo.Echo, tododb *TodoDB) {
	s := &server{tododb: tododb}
	e.GET("/", s.Todos)

	api := e.Group("/api")

	api.GET("/todos", s.GetTodos)
	api.POST("/todos", s.CreateTodo)
	api.DELETE("/todos/:id", s.DeleteTodo)
	api.PUT("/todos/:id/completed", s.CompletedTodo)
}

func (s *server) Todos(c echo.Context) error {
	return util.RenderData(c, "main", "Todos", nil)
}

func (s *server) GetTodos(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			util.RenderError(c, err)
		}
	}()

	todos, err := s.tododb.GetTodos()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "todos", todos)
}

func (s *server) CreateTodo(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			util.RenderError(c, err)
		}
	}()

	name := c.FormValue("name")
	todo, err := s.tododb.InsertTodo(name)
	if err != nil {
		return err
	}

	return c.Render(http.StatusCreated, "todo", todo)
}

func (s *server) DeleteTodo(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			util.RenderError(c, err)
		}
	}()

	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	rows, err := s.tododb.DeleteTodoByID(idInt)
	if err != nil {
		return err
	}

	if rows == 0 {
		return util.RenderError(c, errors.New("todo not found"))
	}

	return c.NoContent(http.StatusOK)
}

func (s *server) CompletedTodo(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			util.RenderError(c, err)
		}
	}()

	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	todo, err := s.tododb.GetTodoByID(idInt)
	if err != nil {
		return err
	}

	todo.Completed = !todo.Completed
	if _, err = s.tododb.UpdateTodo(todo); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "todo", todo)
}
