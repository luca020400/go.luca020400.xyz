package util

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type MainData struct {
	Title string
	Data  interface{}
}

func CreateErrorData(err interface{}) MainData {
	return MainData{
		Title: "Error",
		Data:  err,
	}
}

func RenderError(c echo.Context, err error) error {
	return c.Render(http.StatusInternalServerError, "error", CreateErrorData(err))
}

func CreateData(title string, data interface{}) MainData {
	return MainData{
		Title: title,
		Data:  data,
	}
}

func RenderData(c echo.Context, template string, title string, data interface{}) error {
	return c.Render(http.StatusOK, template, CreateData(title, data))
}
