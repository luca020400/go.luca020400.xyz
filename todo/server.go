package todo

import (
	"errors"
	"net/http"
	"server/user"
	"server/util"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/michaeljs1990/sqlitestore"
)

type server struct {
	tododb *TodoDB
	store  *sqlitestore.SqliteStore
}

func RegisterHandlers(e *echo.Echo, tododb *TodoDB, store *sqlitestore.SqliteStore) {
	s := &server{
		tododb: tododb,
		store:  store,
	}

	e.GET("/", s.Todos)

	api := e.Group("/api")

	api.GET("/todos", s.GetTodos)
	api.POST("/todos", s.CreateTodo)
	api.DELETE("/todos/:id", s.DeleteTodo)
	api.PUT("/todos/:id/completed", s.CompletedTodo)
}

func (s *server) Todos(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			util.RenderError(c, err)
		}
	}()

	session, err := s.store.Get(c.Request(), "session")
	if err != nil {
		return err
	}
	u := session.Values["user"]
	if u == nil {
		return c.Redirect(http.StatusFound, "/login")
	}

	return util.RenderData(c, "main", "Todos", struct {
		User *user.User
	}{
		User: session.Values["user"].(*user.User),
	})
}

func (s *server) GetTodos(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			util.RenderError(c, err)
		}
	}()

	session, err := s.store.Get(c.Request(), "session")
	if err != nil {
		return err
	}
	user := session.Values["user"].(*user.User)

	todos, err := s.tododb.GetTodos(user.ID)
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

	session, err := s.store.Get(c.Request(), "session")
	if err != nil {
		return err
	}
	user := session.Values["user"].(*user.User)

	name := c.FormValue("name")
	todo, err := s.tododb.InsertTodo(user.ID, name)
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

	session, err := s.store.Get(c.Request(), "session")
	if err != nil {
		return err
	}
	user := session.Values["user"].(*user.User)

	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	rows, err := s.tododb.DeleteTodoByID(user.ID, idInt)
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

	session, err := s.store.Get(c.Request(), "session")
	if err != nil {
		return err
	}
	user := session.Values["user"].(*user.User)

	id := c.Param("id")
	idInt, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		return err
	}

	todo, err := s.tododb.GetTodoByID(user.ID, idInt)
	if err != nil {
		return err
	}

	todo.Completed = !todo.Completed
	if _, err = s.tododb.UpdateTodo(todo.Name, todo.Completed, todo.UserID, todo.ID); err != nil {
		return err
	}

	return c.Render(http.StatusOK, "todo", todo)
}
