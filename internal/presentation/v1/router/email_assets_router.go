package router

import (
	"io/fs"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterEmailFontRoutes(engine *gin.Engine, files fs.FS) {
	for _, name := range []string{"regular.woff2", "bold.woff2"} {
		engine.Match([]string{http.MethodGet, http.MethodHead}, "/assets/email/fonts/"+name, func(c *gin.Context) {
			c.Header("Content-Type", "font/woff2")
			c.Header("Cache-Control", "public, max-age=86400")
			c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Credentials", "")
			c.FileFromFS(name, http.FS(files))
		})
	}
}
