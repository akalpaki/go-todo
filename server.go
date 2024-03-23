package main

import (
	"log"
	"log/slog"
	"net/http"
	"strconv"
)

type apiFunc func(http.ResponseWriter, *http.Request) *apiError

func (s *restServer) makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			s.logger.ErrorContext(r.Context(), err.Msg, "Error", err.err)
			http.Error(w, err.Msg, err.Status)
			return
		}
	}
}

type restServer struct {
	listenAddr string
	logger     *slog.Logger
	storer     *Storer
}

func NewRestServer(cfg *config, storer *Storer) *restServer {
	return &restServer{
		listenAddr: cfg.ListenAddr,
		logger:     cfg.Logger,
		storer:     storer,
	}
}

func (s *restServer) Run() {
	http.HandleFunc("GET /", s.makeHTTPHandleFunc(s.handleGetTodos))
	http.HandleFunc("POST /todos", s.makeHTTPHandleFunc(s.handleCreateTodo))
	http.HandleFunc("PUT /todos/{id}", s.makeHTTPHandleFunc(s.handleUpdateTodo))
	http.HandleFunc("GET /todos/{id}", s.makeHTTPHandleFunc(s.handleGetTodo))
	http.HandleFunc("DELETE /todos/{id}", s.makeHTTPHandleFunc(s.handleDeleteTodo))
	http.HandleFunc("GET /todos/{id}/items/{item_id}", s.makeHTTPHandleFunc(s.handleDeleteTodoItem))

	log.Printf("server listening on port: %s", s.listenAddr)

	log.Fatal(http.ListenAndServe(s.listenAddr, nil))
}

func (s *restServer) handleGetTodos(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	resp, err := s.storer.GetTodos(ctx)
	if err != nil {
		return NewAPIError(http.StatusInternalServerError, "failed to retrieve todo lists", err)
	}

	if err := WriteJSON(w, http.StatusOK, resp); err != nil {
		return NewAPIError(http.StatusInternalServerError, "internal server error", err)
	}

	return nil
}

func (s *restServer) handleGetTodo(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return NewAPIError(http.StatusBadRequest, "invalid todo list id", err)
	}

	resp, err := s.storer.GetTodo(ctx, id)
	if err != nil {
		return NewAPIError(http.StatusInternalServerError, "failed to retrieve todo lists", err)
	}

	if err := WriteJSON(w, http.StatusOK, resp); err != nil {
		return NewAPIError(http.StatusInternalServerError, "internal server error", err)
	}

	return nil
}

func (s *restServer) handleCreateTodo(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	var todo Todo

	if err := ReadJSON(w, r, todo); err != nil {
		return NewAPIError(http.StatusBadRequest, "invalid data", err)
	}

	todo, err := s.storer.CreateTodo(ctx, todo)
	if err != nil {
		return NewAPIError(http.StatusInternalServerError, "failed to create todo list", err)
	}

	if err := WriteJSON(w, http.StatusOK, todo); err != nil {
		return NewAPIError(http.StatusInternalServerError, "failed to create todo list", err)
	}

	return nil
}

func (s *restServer) handleUpdateTodo(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return NewAPIError(http.StatusBadRequest, "invalid todo list id", err)
	}

	var updated Todo
	if err := ReadJSON(w, r, updated); err != nil {
		return NewAPIError(http.StatusBadRequest, "invalid data", err)
	}

	res, err := s.storer.UpdateTodo(ctx, id, updated)
	if err != nil {
		return NewAPIError(http.StatusBadRequest, "failed to update todo list", err)
	}

	if err := WriteJSON(w, http.StatusOK, res); err != nil {
		return NewAPIError(http.StatusBadRequest, "internal server error", err)
	}

	return nil
}

func (s *restServer) handleDeleteTodo(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return NewAPIError(http.StatusBadRequest, "invalid todo list id", err)
	}

	if err := s.storer.DeleteTodo(ctx, id); err != nil {
		return NewAPIError(http.StatusInternalServerError, "unable to delete todo list", err)
	}

	if err := WriteJSON(w, http.StatusOK, nil); err != nil {
		return NewAPIError(http.StatusInternalServerError, "internal server error", err)
	}

	return nil
}

func (s *restServer) handleDeleteTodoItem(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("item_id"))
	if err != nil {
		return NewAPIError(http.StatusBadRequest, "invalid todo list id", err)
	}

	if err := s.storer.DeleteTodoItem(ctx, id); err != nil {
		return NewAPIError(http.StatusInternalServerError, "unable to delete todo", err)
	}

	if err := WriteJSON(w, http.StatusOK, nil); err != nil {
		return NewAPIError(http.StatusInternalServerError, "internal server error", err)
	}

	return nil
}
