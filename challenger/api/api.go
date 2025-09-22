package api

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

type user struct {
	FirstName string
	LastName  string
	Biography string
}

func sendJSON(w http.ResponseWriter, resp Response, status int) {
	w.Header().Set("Content-Type", "application/json")
	data, err := json.Marshal(resp)
	if err != nil {
		slog.Error("failed to marshal json data", "error", err)
		sendJSON(
			w,
			Response{Error: "something went wrong"},
			http.StatusInternalServerError,
		)
		return
	}

	w.WriteHeader(status)
	if _, err := w.Write(data); err != nil {
		slog.Error("failed to write json data", "error", err)
		return
	}
}

func NewHandler(db map[string]string) http.Handler {
	r := chi.NewMux()

	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)

	r.Post("/api/users", HandleInsert(db))
	r.Get("/api/users", HandleFindAll(db))
	r.Get("/api/users/{id}", HandleFindById(db))
	r.Delete("/api/users/{id}", HandleDelete(db))
	r.Put("/api/users/{id}", HandleUpdate(db))

	return r
}

type PostBody struct {
	URL       string `json:"url"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Biography string `json:"biography"`
}

type Response struct {
	Error string `json:"error,omitempty"`
	Data  any    `json:"data,omitempty"`
}

func HandleFindAll(db map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if db == nil {
			sendJSON(w, Response{Error: "The users information could not be retrieved"}, http.StatusInternalServerError)
			return
		}
		if len(db) == 0 {
			sendJSON(w, Response{Data: []interface{}{}}, http.StatusOK)
			return
		}

		sendJSON(w, Response{Data: db}, http.StatusOK)
	}
}

func HandleFindById(db map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		if db == nil {
			sendJSON(w, Response{Error: "The user information could not be retrieved"}, http.StatusInternalServerError)
			return
		}

		if user, exists := db[id]; exists {
			sendJSON(w, Response{Data: user}, http.StatusOK)
		} else {
			sendJSON(w, Response{Error: "The user with the specified ID does not exist"}, http.StatusNotFound)
		}

	}
}

func HandleInsert(db map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body PostBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			sendJSON(w, Response{Error: "invalid body"}, http.StatusUnprocessableEntity)
			return
		}
		if body.FirstName == "" || body.LastName == "" || body.Biography == "" {
			sendJSON(w, Response{Error: "Please provide FirstName LastName and bio for the user"}, http.StatusBadRequest)
			return
		}

		newID := uuid.New().String()
		userData := user{FirstName: body.FirstName, LastName: body.LastName, Biography: body.Biography}
		userJSON, err := json.Marshal(userData)

		if err != nil {
			sendJSON(w, Response{Error: "There was an error while saving the user to the database"}, http.StatusInternalServerError)
			return
		}

		db[newID] = string(userJSON)

		if db[newID] == "" {
			sendJSON(w, Response{Error: "There was an error while saving the user to the database"}, http.StatusInternalServerError)
			return
		}

		sendJSON(w, Response{Data: map[string]string{"id": newID}}, http.StatusCreated)
	}
}

func HandleUpdate(db map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")
		var body PostBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			sendJSON(w, Response{Error: "invalid body"}, http.StatusUnprocessableEntity)
			return
		}
		if body.FirstName == "" || body.LastName == "" || body.Biography == "" {
			sendJSON(w, Response{Error: "Please provide FirstName LastName and bio for the user PUTERROR"}, http.StatusBadRequest)
			return
		}
		if _, exists := db[id]; exists {
			updatedUser := user{FirstName: body.FirstName, LastName: body.LastName, Biography: body.Biography}
			userJSON, err := json.Marshal(updatedUser)
			if err != nil {
				sendJSON(w, Response{Error: "The user information could not be modified"}, http.StatusInternalServerError)
				return
			}
			db[id] = string(userJSON)
			if db[id] == "" {
				sendJSON(w, Response{Error: "The user information could not be modified"}, http.StatusInternalServerError)
				return
			}
			sendJSON(w, Response{Data: map[string]string{"id": id}}, http.StatusOK)
		} else {
			sendJSON(w, Response{Error: "he user with the specified ID does not exist"}, http.StatusNotFound)
		}
	}
}
func HandleDelete(db map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := chi.URLParam(r, "id")

		if user, exists := db[id]; exists {
			delete(db, id)
			if _, stillExists := db[id]; stillExists {
				sendJSON(w, Response{Error: "The user with the specified ID does not exist"}, http.StatusInternalServerError)
				return
			}
			sendJSON(w, Response{Data: user}, http.StatusOK)
		} else {
			sendJSON(w, Response{Error: "user not found"}, http.StatusNotFound)
		}
	}
}
