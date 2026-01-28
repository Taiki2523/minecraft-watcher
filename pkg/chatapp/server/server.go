package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gorilla/sessions"

	"github.com/taiki2523/minecraft-watcher/pkg/chatapp/auth"
	"github.com/taiki2523/minecraft-watcher/pkg/chatapp/config"
	"github.com/taiki2523/minecraft-watcher/pkg/chatapp/handlers"
	"github.com/taiki2523/minecraft-watcher/pkg/chatapp/storage"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	store, err := storage.New(cfg.DatabaseDSN)
	if err != nil {
		return err
	}
	defer store.Close()

	sessionStore := sessions.NewCookieStore([]byte(cfg.SessionKey))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,
		HttpOnly: true,
	}

	handler := &handlers.Handler{
		Store:          store,
		Sessions:       sessionStore,
		Google:         auth.NewGoogleClient(cfg.GoogleClientID, cfg.GoogleClientSecret, cfg.GoogleRedirectURL),
		FrontendOrigin: cfg.FrontendOrigin,
	}

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	if exists(cfg.StaticDir) {
		staticPath := http.Dir(cfg.StaticDir)
		mux.Handle("/", spaHandler(staticPath))
	} else {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("chat app server running"))
		})
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	return http.ListenAndServe(addr, mux)
}

func spaHandler(fs http.FileSystem) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := filepath.Clean(r.URL.Path)
		if path == "/" {
			path = "/index.html"
		}

		file, err := fs.Open(path)
		if err != nil {
			fallback, fallbackErr := fs.Open("/index.html")
			if fallbackErr != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			defer fallback.Close()
			http.ServeContent(w, r, "index.html", modTime(fallback), fallback)
			return
		}
		defer file.Close()
		http.ServeContent(w, r, path, modTime(file), file)
	})
}

func modTime(file http.File) time.Time {
	info, err := file.Stat()
	if err != nil {
		return time.Time{}
	}
	return info.ModTime()
}

func exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
