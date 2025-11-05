package routes

import (
	"github.com/gofiber/fiber/v2"
)

func SetupDocs(app *fiber.App) {
	app.Get("/openapi.yaml", func(c *fiber.Ctx) error {
		return c.SendFile("openapi.yaml", true)
	})

	app.Get("/docs", func(c *fiber.Ctx) error {
		html := `<!doctype html>
<html>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>VidNotes API Docs</title>
    <style>body{margin:0;padding:0}</style>
  </head>
  <body>
    <redoc spec-url='/openapi.yaml'></redoc>
    <script src="https://cdn.redoc.ly/redoc/latest/bundles/redoc.standalone.js"></script>
  </body>
</html>`
		c.Type("html")
		return c.SendString(html)
	})
}
