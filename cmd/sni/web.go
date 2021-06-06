package main

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"
)

type WebServer struct {
	listenAddr string

	mux *http.ServeMux

	// broadcast channel to all sockets:
	q chan ViewModelUpdate
}

type ViewModelUpdate struct {
	View      string      `json:"v"`
	ViewModel interface{} `json:"m"`
}

// starts a web server
func NewWebServer(listenAddr string) *WebServer {
	s := &WebServer{
		listenAddr: listenAddr,
		mux:        http.NewServeMux(),
		q:          make(chan ViewModelUpdate, 10),
	}

	// access log file:
	s.mux.Handle("/log.txt", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, logFileName := filepath.Split(logPath)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", logFileName))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		http.ServeFile(w, r, logPath)
	}))

	// serve static content from go-bindata:
	//s.mux.Handle("/", MaxAge(http.FileServer(http.FS(dist.Content))))

	return s
}

func MaxAge(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var age time.Duration
		ext := filepath.Ext(r.URL.String())

		switch ext {
		case ".css", ".js":
			age = (time.Hour * 24 * 30) / time.Second
		case ".jpg", ".jpeg", ".gif", ".png", ".ico", ".cur", ".gz", ".svg", ".svgz",
			".ttf", ".otf",
			".mp4", ".ogg", ".ogv", ".webm", ".htc":
			age = (time.Hour * 24 * 365) / time.Second
		default:
			age = 0
		}

		if age > 0 {
			w.Header().Add("Cache-Control", fmt.Sprintf("max-age=%d, public, must-revalidate, proxy-revalidate", age))
		}

		h.ServeHTTP(w, r)
	})
}

func (s *WebServer) Serve() error {
	// start server:
	return http.ListenAndServe(s.listenAddr, s.mux)
}
