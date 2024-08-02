package main

import (
	"html/template"
	"io"
	"server/todo"
	"server/user"
	"server/util"

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

	// Template
	util.InitTypes()
	t := &Template{
		templates: template.Must(template.ParseGlob("public/views/*.html")),
	}

	e.Renderer = t

	// Routes
	todo.RegisterHandlers(e, tododb, store)
	user.RegisterHandlers(e, usersdb, store)

	e.Static("/static", "public/assets")
	e.Logger.Fatal(e.Start(":1323"))
}
