package gzip

import (
	"compress/gzip"
	"net/http"
	"path"
	"strings"

	"github.com/NYTimes/gziphandler"
)

var allowedTypes = map[string]bool{
	".html":  true,
	".js":    true,
	".css":   true,
	".map":   true,
	".png":   true,
	".jpg":   true,
	".jpeg":  true,
	".gif":   true,
	".woff":  true,
	".woff2": true,
}

var Default = New(gzip.DefaultCompression)

func New(lvl int) func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
	handler := gziphandler.MustNewGzipLevelHandler(lvl)

	return func(rw http.ResponseWriter, req *http.Request, next http.HandlerFunc) {
		ext := strings.ToLower(path.Ext(req.URL.Path))
		if allowedTypes[ext] {
			handle := handler(next)
			handle.ServeHTTP(rw, req)
			return
		}
		next(rw, req)
	}
}
