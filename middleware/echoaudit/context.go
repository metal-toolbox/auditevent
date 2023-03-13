package echoaudit

import (
	"github.com/labstack/echo/v4"
)

// getString returns an echo context value as a string. if the
// value is missing or not a string, it returns an empty string ("").
func getString(c echo.Context, key string) string {
	v := c.Get(key)

	if s, ok := v.(string); ok {
		return s
	}

	return ""
}
