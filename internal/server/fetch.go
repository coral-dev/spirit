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
	"context"
	"database/sql"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/lukewhrit/spacebin/internal/config"
	"github.com/lukewhrit/spacebin/internal/database"
	"github.com/lukewhrit/spacebin/internal/util"
	"golang.org/x/exp/slices"
)

func getDocument(s *Server, ctx context.Context, id string) (database.Document, error) {
	return s.Database.GetDocument(ctx, id)
}

func (s *Server) StaticDocument(w http.ResponseWriter, r *http.Request) {
	params := strings.Split(chi.URLParam(r, "document"), ".")
	id := params[0]

	// Validate document ID
	if len(id) != s.Config.IDLength && !slices.Contains(s.Config.Documents, id) {
		err := fmt.Errorf("id is of length %d, should be %d", len(id), s.Config.IDLength)
		util.RenderError(&resources, w, http.StatusBadRequest, err)
		return
	}

	// Retrieve document from the database
	document, err := getDocument(s, r.Context(), id)

	if err != nil {
		// If the document is not found (ErrNoRows), return the error with a 404
		if errors.Is(err, sql.ErrNoRows) {
			util.RenderError(&resources, w, http.StatusNotFound, err)
			return
		}

		// Otherwise, return the error with a 500
		util.RenderError(&resources, w, http.StatusInternalServerError, err)
		return
	}

	// Reader mode or code mode?
	if r.URL.Query().Get("reader") == "true" {
		t, err := template.ParseFS(resources, "web/reader.html")

		if err != nil {
			util.RenderError(&resources, w, http.StatusInternalServerError, err)
			return
		}

		content := util.ParseMarkdown([]byte(document.Content))

		data := map[string]interface{}{
			"Content":   template.HTML(string(content)),
			"Analytics": template.HTML(config.Config.Analytics),
		}

		if err := t.Execute(w, data); err != nil {
			util.RenderError(&resources, w, http.StatusInternalServerError, err)
			return
		}
	} else {
		t, err := template.ParseFS(resources, "web/document.html")

		if err != nil {
			util.RenderError(&resources, w, http.StatusInternalServerError, err)
			return
		}

		extension := ""

		if len(params) == 2 {
			extension = params[1]
		}

		highlighted, css, err := util.Highlight(document.Content, extension)

		if err != nil {
			util.RenderError(&resources, w, http.StatusInternalServerError, err)
			return
		}

		data := map[string]interface{}{
			"Stylesheet":  template.CSS(css),
			"Content":     document.Content,
			"Highlighted": template.HTML(highlighted),
			"Extension":   extension,
			"Analytics":   template.HTML(config.Config.Analytics),
		}

		if err := t.Execute(w, data); err != nil {
			util.RenderError(&resources, w, http.StatusInternalServerError, err)
			return
		}
	}
}

func (s *Server) FetchDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "document")

	// Validate document ID
	if len(id) != s.Config.IDLength && !slices.Contains(s.Config.Documents, id) {
		err := fmt.Errorf("id is of length %d, should be %d", len(id), s.Config.IDLength)
		util.WriteError(w, http.StatusBadRequest, err)
		return
	}

	document, err := getDocument(s, r.Context(), id)

	if err != nil {
		// If the document is not found (ErrNoRows), return the error with a 404
		if errors.Is(err, sql.ErrNoRows) {
			util.WriteError(w, http.StatusNotFound, err)
			return
		}

		// Otherwise, return the error with a 500
		util.WriteError(w, http.StatusInternalServerError, err)
		return
	}

	// Try responding with the document and a 200, or write an error if that fails
	if err := util.WriteJSON(w, http.StatusOK, document); err != nil {
		util.WriteError(w, http.StatusInternalServerError, err)
		return
	}
}

func (s *Server) FetchRawDocument(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "document")

	// Validate document ID
	if len(id) != s.Config.IDLength && !slices.Contains(s.Config.Documents, id) {
		err := fmt.Errorf("id is of length %d, should be %d", len(id), s.Config.IDLength)
		util.WriteError(w, http.StatusBadRequest, err)
		return
	}

	document, err := getDocument(s, r.Context(), id)

	w.Header().Set("Content-Type", "text/plain")

	if err != nil {
		// If the document is not found (ErrNoRows), return the error with a 404
		if errors.Is(err, sql.ErrNoRows) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(fmt.Sprintf("Document with ID %s not found: %s", id, err.Error())))
			return
		}

		// Otherwise, return the error with a 500
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("Error fetching document with ID %s: %s", id, err.Error())))
		return
	}

	// Respond with only the documents content
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(document.Content))
}
