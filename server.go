package main

import (
	"log"
	"log/slog"
	"net/http"
	"strconv"
)

type apiFunc func(http.ResponseWriter, *http.Request) *apiError

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			slog.ErrorContext(r.Context(), err.Msg, "Error", err.err)
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
	// Todo endpoints: crud on the todo list entity
	http.HandleFunc("GET /v1/todos", makeHTTPHandleFunc(s.handleGetTodos))
	http.HandleFunc("POST /v1/todos", makeHTTPHandleFunc(s.handleCreateTodo))
	http.HandleFunc("GET /v1/todos/{id}", makeHTTPHandleFunc(s.handleGetTodo))
	http.HandleFunc("PUT /v1/todos/{id}", makeHTTPHandleFunc(s.handleUpdateTodo))
	http.HandleFunc("DELETE /v1/todos/{id}", makeHTTPHandleFunc(s.handleDeleteTodo))

	// Todo item endpoints: crud on todo list items
	http.HandleFunc("GET /v1/todos/{id}/items", makeHTTPHandleFunc(s.handleGetTodoItems))
	http.HandleFunc("POST /v1/todos/{id}/items", makeHTTPHandleFunc(s.handleAddTodoItem))
	http.HandleFunc("PUT /v1/todos/{id}/items/{itemNo}", makeHTTPHandleFunc(s.handleEditTodoItem))
	http.HandleFunc("DELETE /v1/todos/{id}/items/{itemNo}", makeHTTPHandleFunc(s.handleDeleteTodoItem))

	log.Printf("server listening on port: %s", s.listenAddr)

	log.Fatal(http.ListenAndServe(s.listenAddr, nil))
}

func (s *restServer) handleGetTodos(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	resp, err := s.storer.GetTodos(ctx)
	if err != nil {
		return internalErrorResponse("failed tor retrieve todo lists", err)
	}

	if err := WriteJSON(w, http.StatusOK, resp); err != nil {
		return internalErrorResponse("failed to process request", err) // FIXME: this error is wrong, I put it as standin for now
	}

	return nil
}

func (s *restServer) handleGetTodo(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponse("invalid todo list id", err)
	}

	resp, err := s.storer.GetTodo(ctx, id)
	if err != nil {
		return internalErrorResponse("failed to retrieve todo list", err)
	}

	if err := WriteJSON(w, http.StatusOK, resp); err != nil {
		return internalErrorResponse("failed to process request", err)
	}

	return nil
}

func (s *restServer) handleCreateTodo(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	var todo CreateTodo

	if err := ReadJSON(w, r, &todo); err != nil {
		return badRequestResponse("invalid data", err)
	}

	out, err := s.storer.CreateTodo(ctx, todo)
	if err != nil {
		return internalErrorResponse("failed to create todo list", err)
	}

	if err := WriteJSON(w, http.StatusOK, out); err != nil {
		return internalErrorResponse("failed to process request", err)
	}

	return nil
}

func (s *restServer) handleUpdateTodo(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	var update UpdateTodo

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponse("invalid todo list id", err)
	}

	if err := ReadJSON(w, r, &update); err != nil {
		return badRequestResponse("invalid data", err)
	}

	err = s.storer.UpdateTodo(ctx, id, update)
	if err != nil {
		return internalErrorResponse("failed to update todo list", err)
	}

	if err := WriteJSON(w, http.StatusOK, nil); err != nil {
		return internalErrorResponse("unable to process request", err)
	}

	return nil
}

func (s *restServer) handleDeleteTodo(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponse("invalid todo list id", err)
	}

	if err := s.storer.DeleteTodo(ctx, id); err != nil {
		return internalErrorResponse("failed to delete todo list", err)
	}

	if err := WriteJSON(w, http.StatusOK, nil); err != nil {
		return internalErrorResponse("failed to process request", err)
	}

	return nil
}

func (s *restServer) handleGetTodoItems(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	todoID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponse("invalid todo list id", err)
	}

	items, err := s.storer.GetTodoItems(ctx, todoID)
	if err != nil {
		return internalErrorResponse("failed to retrieve todo items", err)
	}

	if err := WriteJSON(w, http.StatusOK, items); err != nil {
		return internalErrorResponse("failed to process request", err)
	}

	return nil
}

func (s *restServer) handleAddTodoItem(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	var item Item

	if err := ReadJSON(w, r, &item); err != nil {
		return badRequestResponse("invalid data", err)
	}

	if err := s.storer.AddTodoItem(ctx, item); err != nil {
		return internalErrorResponse("failed to add todo item", err)
	}

	return nil
}

func (s *restServer) handleEditTodoItem(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	var item Item

	if err := ReadJSON(w, r, &item); err != nil {
		return badRequestResponse("invalid data", err)
	}

	if err := s.storer.UpdateTodoItem(ctx, item); err != nil {
		return internalErrorResponse("failed to add todo item", err)
	}

	return nil
}

func (s *restServer) handleDeleteTodoItem(w http.ResponseWriter, r *http.Request) *apiError {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("itemNo"))
	if err != nil {
		return badRequestResponse("invalid item number", err)
	}

	if err := s.storer.DeleteTodoItem(ctx, id); err != nil {
		return internalErrorResponse("failed to delete todo list item", err)
	}

	if err := WriteJSON(w, http.StatusOK, nil); err != nil {
		return internalErrorResponse("failed to process request", err)
	}

	return nil
}
