package in

import (
	"app-mobile-downloader/internal/shared/server"

	"github.com/Ignaciojeria/ioc"
	"github.com/go-fuego/fuego"
)

var _ = ioc.Register(helloWorldHandler)

func helloWorldHandler(s *server.Server) {
	fuego.All(s.Server, "/", func(c fuego.ContextNoBody) (string, error) {
		return "Te puse el editor 2!", nil
	})
}
