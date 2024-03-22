package main

import (
	"log"
	"log/slog"
	"net/http"
)

type restServer struct {
	listenAddr string
	logger     *slog.Logger
	storer     *Storer
}

func NewRestServer(cfg config, storer *Storer) *restServer {
	return &restServer{
		listenAddr: cfg.ListenAddr,
		logger:     cfg.Logger,
		storer:     storer,
	}
}

func (s *restServer) Run() {
	http.HandleFunc("GET /", s.handleCreateTodo)

	log.Printf("server listening on port: %s", s.listenAddr)

	log.Fatal(http.ListenAndServe(s.listenAddr, nil))
}

func (s *restServer) handleCreateTodo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var todo Todo

	if err := ReadJSON(w, r, todo); err != nil {
		s.logger.ErrorContext(ctx, "createTodo", slog.Any("error", err))
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	todo, err := s.storer.CreateTodo(ctx, todo)
	if err != nil {
		s.logger.ErrorContext(ctx, "createTodo", slog.Any("error", err))
		http.Error(w, "failed to create todo list", http.StatusInternalServerError)
		return
	}

	if err := WriteJSON(w, http.StatusOK, todo); err != nil {
		s.logger.ErrorContext(ctx, "createTodo", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}

func (s *restServer) handleAddTodoItems(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var todo Todo

	if err := ReadJSON(w, r, todo); err != nil {
		s.logger.ErrorContext(ctx, "addTodoItems", slog.Any("error", err))
		http.Error(w, "Invalid data", http.StatusBadRequest)
		return
	}

	items, err := s.storer.AddTodoItems(ctx, todo.Items)
	if err != nil {
		s.logger.ErrorContext(ctx, "addTodoItems", slog.Any("error", err))
		http.Error(w, "failed to add todo items", http.StatusInternalServerError)
		return
	}

	if err := WriteJSON(w, http.StatusOK, items); err != nil {
		s.logger.ErrorContext(ctx, "addTodoItems", slog.Any("error", err))
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
}
