package middleware

import (
	"fmt"

	"github.com/VictoriaMetrics/metrics"
	"github.com/labstack/echo/v4"
)

type Victoria struct{}

func (s *Victoria) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := next(c); err != nil {
			c.Error(err)
		}
		m := fmt.Sprintf(
			`gonah_requests_total{path=%q, method=%q, status="%d"}`,
			c.Request().URL.Path,
			c.Request().Method,
			c.Response().Status,
		)
		metrics.GetOrCreateCounter(m).Inc()
		return nil
	}
}
