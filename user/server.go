package user

import (
	"server/util"

	"github.com/labstack/echo/v4"
	"github.com/michaeljs1990/sqlitestore"
)

type server struct {
	usersdb *UserDB
	store   *sqlitestore.SqliteStore
}

func RegisterHandlers(e *echo.Echo, usersdb *UserDB, store *sqlitestore.SqliteStore) {
	s := &server{
		usersdb: usersdb,
		store:   store,
	}

	e.GET("/login", s.Login)
	e.POST("/login", s.DoLogin)
	e.GET("/register", s.Register)
	e.POST("/register", s.DoRegister)
	e.GET("/logout", s.Logout)
}

func (s *server) Login(c echo.Context) error {
	if _, logged := util.IsLoggedIn(s.store, c); logged {
		return c.Redirect(302, "/")
	}

	return util.RenderData(c, "login", "Login", nil)
}

func (s *server) DoLogin(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			util.RenderError(c, err)
		}
	}()

	nick := c.FormValue("nick")
	password := c.FormValue("password")

	user, hash, err := s.usersdb.GetUserByNick(nick)
	if err != nil {
		return err
	}

	if match, err := util.ComparePasswordAndHash(password, hash); err != nil || !match {
		return c.String(401, "Unauthorized")
	}

	session, err := s.store.New(c.Request(), "session")
	if err != nil {
		return err
	}

	session.Values["user"] = user
	session.Save(c.Request(), c.Response())

	return c.Redirect(302, "/")
}

func (s *server) Register(c echo.Context) error {
	if _, logged := util.IsLoggedIn(s.store, c); logged {
		return c.Redirect(302, "/")
	}

	return util.RenderData(c, "register", "Register", nil)
}

func (s *server) DoRegister(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			util.RenderError(c, err)
		}
	}()

	nick := c.FormValue("nick")
	password := c.FormValue("password")

	hash, err := util.GenerateHashFromPassword(password)
	if err != nil {
		return err
	}

	_, err = s.usersdb.InsertUser(nick, hash)
	if err != nil {
		return err
	}

	return c.Redirect(302, "/login")
}

func (s *server) Logout(c echo.Context) error {
	var err error
	defer func() {
		if err != nil {
			util.RenderError(c, err)
		}
	}()

	if _, logged := util.IsLoggedIn(s.store, c); !logged {
		return c.Redirect(302, "/")
	}

	session, err := s.store.Get(c.Request(), "session")
	if err != nil {
		return err
	}

	session.Options.MaxAge = -1
	session.Save(c.Request(), c.Response())

	return c.Redirect(302, "/")
}
