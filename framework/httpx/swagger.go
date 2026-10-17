// Package httpx provides unified HTTP response and documentation helpers for gorp framework.
package httpx

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/ngq/gorp/framework/contract/support"
)

// SwaggerUIHTML returns a lightweight, self-contained HTML string loading Swagger UI CDN.
func SwaggerUIHTML(specURL, title string) string {
	if title == "" {
		title = "gorp API Documentation"
	}
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="UTF-8">
  <title>%s</title>
  <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    html { box-sizing: border-box; overflow: -moz-scrollbars-vertical; overflow-y: scroll; }
    *, *:before, *:after { box-sizing: inherit; }
    body { margin: 0; background: #fafafa; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    window.onload = function() {
      SwaggerUIBundle({
        url: "%s",
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIBundle.SwaggerUIStandalonePreset
        ],
        layout: "BaseLayout"
      });
    };
  </script>
</body>
</html>`, title, specURL)
}

// GinSwaggerHandler returns a Gin handler that serves interactive Swagger UI or the spec JSON.
func GinSwaggerHandler(specPath string) gin.HandlerFunc {
	return func(c *gin.Context) {
		reqPath := c.Request.URL.Path
		if strings.HasSuffix(reqPath, ".json") || strings.HasSuffix(reqPath, "/spec") {
			targetPath := specPath
			if targetPath == "" {
				targetPath = support.ResolveResourcePath(filepath.Join("docs", "openapi.json"), "docs")
			} else {
				targetPath = support.ResolveResourcePath(targetPath, "docs")
			}
			data, err := os.ReadFile(targetPath)
			if err != nil {
				fallbackPath := support.ResolveResourcePath(filepath.Join("docs", "swagger.json"), "docs")
				var fallbackErr error
				data, fallbackErr = os.ReadFile(fallbackPath)
				if fallbackErr != nil {
					c.JSON(http.StatusNotFound, gin.H{
						"error": fmt.Sprintf("spec file not found: %s", specPath),
					})
					return
				}
			}
			c.Data(http.StatusOK, "application/json; charset=utf-8", data)
			return
		}

		specURL := reqPath
		if !strings.HasSuffix(specURL, "/") {
			specURL += "/"
		}
		specURL += "spec"

		html := SwaggerUIHTML(specURL, "gorp API Documentation")
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(html))
	}
}
