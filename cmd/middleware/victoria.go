package middleware

import (
	"fmt"
	"time"

	"github.com/VictoriaMetrics/metrics"
	"github.com/labstack/echo/v4"
)

type Victoria struct{}

func (s *Victoria) Process(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		md := fmt.Sprintf(
			`gonah_request_duration_seconds{path=%q, method=%q}`,
			c.Request().URL.Path,
			c.Request().Method,
		)
		dur := metrics.GetOrCreateHistogram(md) // todo: cache histogram

		if err := next(c); err != nil {
			c.Error(err)
		}
		dur.UpdateDuration(start)
		re := fmt.Sprintf(
			`gonah_requests_total{path=%q, method=%q, status="%d"}`,
			c.Request().URL.Path,
			c.Request().Method,
			c.Response().Status,
		)
		metrics.GetOrCreateCounter(re).Inc() // todo: cache counter
		return nil
	}
}
