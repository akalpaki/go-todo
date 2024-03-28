package main

import (
	"log"
	"log/slog"
	"net/http"
	"strconv"
)

const DEFAULT_LIMIT = 10

type apiFunc func(http.ResponseWriter, *http.Request) *apiErrorV2

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}
		slog.Error(err.Title, "error", err.Error())
		body, unmarshErr := err.ResponseBody()
		if unmarshErr != nil {
			slog.Error("an error occured", "error", unmarshErr.Error())
		}
		status, headers := err.ResponseHeaders()
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		w.Write(body)
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
	// User endpoints: crud and login of users
	http.HandleFunc("POST /v1/user", makeHTTPHandleFunc())
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

func (s *restServer) handleCreateUser(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var user User
	if err := ReadJSON(w, r, &user); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

}

func (s *restServer) handleGetTodos(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	queryParams := r.URL.Query()

	page, err := strconv.Atoi(queryParams.Get("page"))
	if err != nil {
		page = 0
	}

	limit, err := strconv.Atoi(queryParams.Get("limit"))
	if err != nil {
		limit = DEFAULT_LIMIT
	}

	resp, err := s.storer.GetTodos(ctx, limit, page)
	if err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to retrieve todo lists", err)
		}
	}

	if err := WriteJSON(w, http.StatusOK, resp); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (s *restServer) handleGetTodo(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid todo list id", err)
	}

	resp, err := s.storer.GetTodo(ctx, id)
	if err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to retrieve todo list", err)
		}
	}

	if err := WriteJSON(w, http.StatusOK, resp); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (s *restServer) handleCreateTodo(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var todo CreateTodo

	if err := ReadJSON(w, r, &todo); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	out, err := s.storer.CreateTodo(ctx, todo)
	if err != nil {
		return internalErrorResponseV2("failed to create todo list", err)
	}

	if err := WriteJSON(w, http.StatusOK, out); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (s *restServer) handleUpdateTodo(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var update UpdateTodo

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid todo list id", err)
	}

	if err := ReadJSON(w, r, &update); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	err = s.storer.UpdateTodo(ctx, id, update)
	if err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to update todo list", err)
		}
	}

	if err := WriteJSON(w, http.StatusOK, nil); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (s *restServer) handleDeleteTodo(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid todo list id", err)
	}

	if err := s.storer.DeleteTodo(ctx, id); err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to delete todo list", err)
		}
	}

	if err := WriteJSON(w, http.StatusOK, nil); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (s *restServer) handleGetTodoItems(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	todoID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid todo list id", err)
	}

	items, err := s.storer.GetTodoItems(ctx, todoID)
	if err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to retrieve todo items", err)
		}
	}

	if err := WriteJSON(w, http.StatusOK, items); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (s *restServer) handleAddTodoItem(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var item Item

	if err := ReadJSON(w, r, &item); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	if err := s.storer.AddTodoItem(ctx, item); err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to add todo item", err)
		}
	}

	return nil
}

func (s *restServer) handleEditTodoItem(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var item Item

	if err := ReadJSON(w, r, &item); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	if err := s.storer.UpdateTodoItem(ctx, item); err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to add todo item", err)
		}
	}

	return nil
}

func (s *restServer) handleDeleteTodoItem(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("itemNo"))
	if err != nil {
		return badRequestResponseV2("invalid item number", err)
	}

	if err := s.storer.DeleteTodoItem(ctx, id); err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to delete todo list item", err)
		}
	}

	if err := WriteJSON(w, http.StatusOK, nil); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}
