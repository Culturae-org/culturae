// backend/internal/handler/admin/ui.go

package admin

import (
	"embed"
	"io"
	"io/fs"
	"net/http"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:ui/dist
var dashboardFS embed.FS

var mimeTypes = map[string]string{
	".html":  "text/html",
	".js":    "application/javascript",
	".css":   "text/css",
	".json":  "application/json",
	".png":   "image/png",
	".jpg":   "image/jpeg",
	".jpeg":  "image/jpeg",
	".svg":   "image/svg+xml",
	".ico":   "image/x-icon",
	".woff":  "font/woff",
	".woff2": "font/woff2",
	".ttf":   "font/ttf",
	".eot":   "application/vnd.ms-fontobject",
	".map":   "application/json",
}

func ServeDashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		urlPath := strings.TrimPrefix(c.Request.URL.Path, "/console")

		if urlPath == "" || urlPath == "/" {
			serveDashboardFile(c, "index.html")
			return
		}

		filePath := strings.TrimPrefix(urlPath, "/")

		f, err := dashboardFS.Open("ui/dist/" + filePath)
		if err == nil {
			defer func() {
				_ = f.Close()
			}()
			stat, err := f.Stat()
			if err == nil && !stat.IsDir() {
				if rs, ok := f.(io.ReadSeeker); ok {
					serveWithContentType(c, filePath, stat, rs)
					return
				}
			}
		}

		f, err = dashboardFS.Open("ui/dist/" + filePath + "/index.html")
		if err == nil {
			defer func() {
				_ = f.Close()
			}()
			stat, err := f.Stat()
			if err == nil && !stat.IsDir() {
				if rs, ok := f.(io.ReadSeeker); ok {
					serveWithContentType(c, filePath+"/index.html", stat, rs)
					return
				}
			}
		}

		serveDashboardFile(c, "index.html")
	}
}

func serveWithContentType(c *gin.Context, filePath string, stat fs.FileInfo, r io.ReadSeeker) {
	ext := strings.ToLower(path.Ext(filePath))
	contentType := mimeTypes[ext]
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	c.Header("Content-Type", contentType)
	if ext == ".html" {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	} else {
		c.Header("Cache-Control", "public, max-age=31536000, immutable")
	}
	http.ServeContent(c.Writer, c.Request, stat.Name(), stat.ModTime(), r)
}

func serveDashboardFile(c *gin.Context, name string) {
	f, err := dashboardFS.Open("ui/dist/" + name)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}
	defer func() {
		_ = f.Close()
	}()

	stat, err := f.Stat()
	if err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}

	c.Header("Content-Type", "text/html")
	c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
	if rs, ok := f.(io.ReadSeeker); ok {
		http.ServeContent(c.Writer, c.Request, stat.Name(), stat.ModTime(), rs)
	}
}
