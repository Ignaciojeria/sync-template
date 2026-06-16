package in

import (
	"app-mobile-downloader/internal/shared/infrastructure/test"
	"app-mobile-downloader/templates"

	"github.com/go-fuego/fuego"
)

func renderResultAndDashboard(c fuego.ContextNoBody, state templates.TestRunState) (string, error) {
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	w := c.Response()
	ctx := c.Context()
	if err := test.RenderResultAndDashboard(w, ctx, state); err != nil {
		return "", err
	}
	return "", nil
}
