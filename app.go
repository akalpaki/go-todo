package main

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
)

// |++++++++++++++++++++++++++++++++++++++|
// |              Middleware              |
// |++++++++++++++++++++++++++++++++++++++|

const DEFAULT_LIMIT = 10

type apiFunc func(http.ResponseWriter, *http.Request) *apiErrorV2

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f(w, r)
		if err == nil {
			return
		}
		slog.Error(err.Title, "error", err.Error())
		body, unmarshErr := err.responseBody()
		if unmarshErr != nil {
			slog.Error("an error occured", "error", unmarshErr.Error())
		}
		status, headers := err.responseHeaders()
		for k, v := range headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(status)
		w.Write(body)
	}
}

// |++++++++++++++++++++++++++++++++++++++|
// |          Server and Handlers         |
// |++++++++++++++++++++++++++++++++++++++|

type application struct {
	logger     *slog.Logger
	repository *repository
	handler    *http.ServeMux
}

func newApplication(logger *slog.Logger, repository *repository) *application {
	app := &application{
		logger:     logger,
		repository: repository,
	}
	app.SetupRoutes()
	return app
}

func (a *application) SetupRoutes() {
	mux := http.NewServeMux()
	// User endpoints: crud and login of users
	mux.HandleFunc("POST /v1/user", makeHTTPHandleFunc(a.handleCreateUser))
	mux.HandleFunc("POST /v1/user/login", makeHTTPHandleFunc(a.handleLoginUser))

	// Todo endpoints: crud on the todo list entity
	mux.HandleFunc("POST /v1/todos", makeHTTPHandleFunc(a.handleCreateTodo))
	mux.HandleFunc("GET /v1/todos", withJWTTodoAuth(makeHTTPHandleFunc(a.handleGetTodos), a.repository))
	mux.HandleFunc("GET /v1/todos/{id}", withJWTTodoAuth(makeHTTPHandleFunc(a.handleGetTodo), a.repository))
	mux.HandleFunc("PUT /v1/todos/{id}", withJWTTodoAuth(makeHTTPHandleFunc(a.handleUpdateTodo), a.repository))
	mux.HandleFunc("DELETE /v1/todos/{id}", withJWTTodoAuth(makeHTTPHandleFunc(a.handleDeleteTodo), a.repository))

	// Todo item endpoints: crud on todo list items
	mux.HandleFunc("GET /v1/todos/{id}/items", withJWTTodoAuth(makeHTTPHandleFunc(a.handleGetTodoItems), a.repository))
	mux.HandleFunc("POST /v1/todos/{id}/items", withJWTTodoAuth(makeHTTPHandleFunc(a.handleAddTodoItem), a.repository))
	mux.HandleFunc("PUT /v1/todos/{id}/items/{itemNo}", withJWTTodoAuth(makeHTTPHandleFunc(a.handleEditTodoItem), a.repository))
	mux.HandleFunc("DELETE /v1/todos/{id}/items/{itemNo}", withJWTTodoAuth(makeHTTPHandleFunc(a.handleDeleteTodoItem), a.repository))

	a.handler = mux
}

func (a *application) handleCreateUser(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var user User
	if err := readJSON(w, r, &user); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	validate := validator.New()

	err := validate.Struct(user)
	if err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	registeredUser, err := a.repository.CreateUser(ctx, user)
	if err != nil {
		return internalErrorResponseV2("failed to create user", err)
	}

	if err := writeJSON(w, http.StatusOK, registeredUser); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (a *application) handleLoginUser(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var user User
	if err := readJSON(w, r, &user); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	validate := validator.New()
	err := validate.Struct(user)
	if err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	registeredUser, err := a.repository.GetUserByEmail(ctx, user.Email)
	if err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	if !checkPasswordHash(user.Password, registeredUser.Password) {
		return badRequestResponseV2("wrong password", err)
	}

	token, err := createAccessToken(user.ID)
	if err != nil {
		return internalErrorResponseV2("failed to log in user", err)
	}

	successfulLoginResponse(w, token)
	return nil
}

func (a *application) handleGetTodos(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
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

	userID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	resp, err := a.repository.GetTodosByUserID(ctx, userID, limit, page)
	if err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to retrieve todo lists", err)
		}
	}

	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (a *application) handleGetTodo(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid todo list id", err)
	}

	resp, err := a.repository.GetTodo(ctx, id)
	if err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to retrieve todo list", err)
		}
	}

	if err := writeJSON(w, http.StatusOK, resp); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (a *application) handleCreateTodo(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var todo CreateTodo

	if err := readJSON(w, r, &todo); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	out, err := a.repository.CreateTodo(ctx, todo)
	if err != nil {
		return internalErrorResponseV2("failed to create todo list", err)
	}

	if err := writeJSON(w, http.StatusOK, out); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (a *application) handleUpdateTodo(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var update UpdateTodo

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid todo list id", err)
	}

	if err := readJSON(w, r, &update); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	err = a.repository.UpdateTodo(ctx, id, update)
	if err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to update todo list", err)
		}
	}

	if err := writeJSON(w, http.StatusOK, nil); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (a *application) handleDeleteTodo(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid todo list id", err)
	}

	if err := a.repository.DeleteTodo(ctx, id); err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to delete todo list", err)
		}
	}

	if err := writeJSON(w, http.StatusOK, nil); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (a *application) handleGetTodoItems(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	todoID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid todo list id", err)
	}

	items, err := a.repository.GetTodoItems(ctx, todoID)
	if err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to retrieve todo items", err)
		}
	}

	if err := writeJSON(w, http.StatusOK, items); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (a *application) handleAddTodoItem(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var item Item

	if err := readJSON(w, r, &item); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	if err := a.repository.AddTodoItem(ctx, item); err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to add todo item", err)
		}
	}

	return nil
}

func (a *application) handleEditTodoItem(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var item Item

	if err := readJSON(w, r, &item); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	if err := a.repository.UpdateTodoItem(ctx, item); err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to add todo item", err)
		}
	}

	return nil
}

func (a *application) handleDeleteTodoItem(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("itemNo"))
	if err != nil {
		return badRequestResponseV2("invalid item number", err)
	}

	if err := a.repository.DeleteTodoItem(ctx, id); err != nil {
		switch err {
		case errNotFound:
			return notFoundResponseV2()
		default:
			return internalErrorResponseV2("failed to delete todo list item", err)
		}
	}

	if err := writeJSON(w, http.StatusOK, nil); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}
