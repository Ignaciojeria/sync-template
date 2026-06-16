package in

import (
	"net/http"

	"app-mobile-downloader/internal/shared/access"
	"app-mobile-downloader/internal/shared/server"
	"app-mobile-downloader/internal/shared/server/middleware"
	infratest "app-mobile-downloader/internal/shared/infrastructure/test"
	"app-mobile-downloader/templates"

	"github.com/Ignaciojeria/ioc"
	"github.com/go-fuego/fuego"
)

var _ = ioc.Register(testReportPageHandler)

func testReportPageHandler(s *server.Server) {
	fuego.Get(s.Server, "/report/tests", testReportPage)
}

func testReportPage(c fuego.ContextNoBody) (string, error) {
	claims, ok := middleware.JWTClaimsFromContext(c.Context())
	if !ok {
		return "", fuego.HTTPError{Status: http.StatusUnauthorized, Detail: "unauthorized"}
	}
	email := firstStringClaim(claims, "email")
	if !access.IsAllowedEditorEmail(email) {
		return "", fuego.HTTPError{Status: http.StatusForbidden, Detail: "forbidden"}
	}

	state := infratest.LoadLastRunState()
	c.SetHeader("Content-Type", "text/html; charset=utf-8")
	return "", templates.TestReportPage(state).Render(c.Context(), c.Response())
}
