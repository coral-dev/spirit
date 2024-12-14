/*
 * Copyright 2020-2024 Luke Whritenour

 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at

 *     http://www.apache.org/licenses/LICENSE-2.0

 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
	"text/template"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/go-chi/httprate"
	"github.com/lukewhrit/spacebin/internal/config"
	"github.com/lukewhrit/spacebin/internal/database"
	"github.com/lukewhrit/spacebin/internal/util"
	"github.com/rs/zerolog/log"
)

//go:embed web/*
var resources embed.FS

type Server struct {
	Router   *chi.Mux
	Config   *config.Cfg
	Database database.Database
}

func NewServer(config *config.Cfg, db database.Database) *Server {
	s := &Server{}
	s.Router = chi.NewRouter()
	s.Config = config
	s.Database = db
	return s
}

// serveFiles conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func serveFiles(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}

	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

// These functions should be executed in the order they are defined, that is:
//  1. Mount middleware - MountMiddleware()
//  2. Add security headers - RegisterHeaders()
//  3. Load static content, if enabled - MountStatic()
//  4. Mount API routes - MountHandlers()

func (s *Server) MountMiddleware() {
	// Register middleware
	s.Router.Use(util.Logger)
	s.Router.Use(middleware.RequestID)
	s.Router.Use(middleware.RealIP)
	s.Router.Use(middleware.AllowContentType("application/json", "multipart/form-data"))

	// Ratelimiter
	reqs, per, err := util.ParseRatelimiterString(s.Config.Ratelimiter)

	if err != nil {
		log.Error().
			Err(err).
			Msg("Parse Ratelimiter Error")
	}

	s.Router.Use(httprate.LimitAll(reqs, per))
	s.Router.Use(middleware.Heartbeat("/ping"))
	s.Router.Use(middleware.Recoverer)

	// CORS
	s.Router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))

	// Basic Auth
	if s.Config.Username != "" && s.Config.Password != "" {
		s.Router.Use(middleware.BasicAuth("spacebin", map[string]string{
			s.Config.Username: s.Config.Password,
		}))
	}
}

func (s *Server) RegisterHeaders() {
	s.Router.Use(middleware.SetHeader("X-Download-Options", "noopen"))
	s.Router.Use(middleware.SetHeader("X-DNS-Prefetch-Control", "off"))
	s.Router.Use(middleware.SetHeader("X-Frame-Options", "SAMEORIGIN"))
	s.Router.Use(middleware.SetHeader("X-XSS-Protection", "1; mode=block"))
	s.Router.Use(middleware.SetHeader("X-Content-Type-Options", "nosniff"))
	s.Router.Use(middleware.SetHeader("Referrer-Policy", "no-referrer-when-downgrade"))
	s.Router.Use(middleware.SetHeader("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload"))
	s.Router.Use(middleware.SetHeader("Content-Security-Policy", s.Config.ContentSecurityPolicy))
}

func (s *Server) MountStatic() {
	// Static content views and homepage
	filesDir, err := fs.Sub(resources, "web/static")

	if err != nil {
		log.Error().
			Err(err).
			Msg("Error loading static files")
	}

	serveFiles(s.Router, "/static/", http.FS(filesDir))

	s.Router.Get("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		file, err := resources.ReadFile("web/static/robots.txt")

		if err != nil {
			util.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		w.Write(file)
	})

	s.Router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		t, err := template.ParseFS(resources, "web/index.html")

		if err != nil {
			util.WriteError(w, http.StatusInternalServerError, err)
			return
		}

		err = t.Execute(w, map[string]interface{}{
			"Analytics": config.Config.Analytics,
		})

		if err != nil {
			util.WriteError(w, http.StatusInternalServerError, err)
			return
		}
	})
}

func (s *Server) MountHandlers() {
	// Register routes
	s.Router.Get("/config", s.GetConfig)

	// Document routes
	s.Router.Post("/api/", s.CreateDocument)
	s.Router.Get("/api/{document}", s.FetchDocument)
	s.Router.Get("/api/{document}/raw", s.FetchRawDocument)

	// Account routes
	s.Router.Post("/api/signin", s.SignIn)
	s.Router.Post("/api/signup", s.SignUp)

	// Static routes
	s.Router.Post("/", s.StaticCreateDocument)
	s.Router.Get("/{document}", s.StaticDocument)
	s.Router.Get("/{document}/raw", s.FetchRawDocument)

	// Legacy routes
	s.Router.Post("/v1/documents/", s.CreateDocument)
	s.Router.Get("/v1/documents/{document}", s.FetchDocument)
	s.Router.Get("/v1/documents/{document}/raw", s.FetchRawDocument)
}
