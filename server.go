package main

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
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

// withJWTTodoAuth is authentication middleware for routes handling the todo resource.
func withJWTTodoAuth(handlerFunc http.HandlerFunc, s *Storer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("x-jwt-token")
		if tokenStr == "" {
			WriteJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		token, err := validateJWT(tokenStr)
		if err != nil {
			WriteJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		todoID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			WriteJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		todo, err := s.GetTodoMetadataByID(todoID)
		if err != nil {
			WriteJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		if todo.UserID != int(claims["sub"].(float64)) {
			WriteJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		if claims["iss"] != "todo" {
			WriteJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		if claims["aud"] != "todo" {
			WriteJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}
		exp, err := claims.GetExpirationTime()
		if err != nil || exp == nil {
			WriteJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		if isExpired(exp.Time) {
			WriteJSON(w, http.StatusForbidden, apiErrorV2{
				Type:   errTypeForbidden,
				Title:  "Access Denied",
				Status: http.StatusForbidden,
			})
			return
		}

		handlerFunc(w, r)
	}
}

// |++++++++++++++++++++++++++++++++++++++|
// |          Server and Handlers         |
// |++++++++++++++++++++++++++++++++++++++|

type application struct {
	logger  *slog.Logger
	storer  *Storer
	handler *http.ServeMux
}

func NewApplication(logger *slog.Logger, storer *Storer) *application {
	app := &application{
		logger: logger,
		storer: storer,
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
	mux.HandleFunc("GET /v1/todos", withJWTTodoAuth(makeHTTPHandleFunc(a.handleGetTodos), a.storer))
	mux.HandleFunc("GET /v1/todos/{id}", withJWTTodoAuth(makeHTTPHandleFunc(a.handleGetTodo), a.storer))
	mux.HandleFunc("PUT /v1/todos/{id}", withJWTTodoAuth(makeHTTPHandleFunc(a.handleUpdateTodo), a.storer))
	mux.HandleFunc("DELETE /v1/todos/{id}", withJWTTodoAuth(makeHTTPHandleFunc(a.handleDeleteTodo), a.storer))

	// Todo item endpoints: crud on todo list items
	mux.HandleFunc("GET /v1/todos/{id}/items", withJWTTodoAuth(makeHTTPHandleFunc(a.handleGetTodoItems), a.storer))
	mux.HandleFunc("POST /v1/todos/{id}/items", withJWTTodoAuth(makeHTTPHandleFunc(a.handleAddTodoItem), a.storer))
	mux.HandleFunc("PUT /v1/todos/{id}/items/{itemNo}", withJWTTodoAuth(makeHTTPHandleFunc(a.handleEditTodoItem), a.storer))
	mux.HandleFunc("DELETE /v1/todos/{id}/items/{itemNo}", withJWTTodoAuth(makeHTTPHandleFunc(a.handleDeleteTodoItem), a.storer))

	a.handler = mux
}

func (a *application) handleCreateUser(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var user User
	if err := ReadJSON(w, r, &user); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	validate := validator.New()

	err := validate.Struct(user)
	if err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	registeredUser, err := a.storer.CreateUser(ctx, user)
	if err != nil {
		return internalErrorResponseV2("failed to create user", err)
	}

	if err := WriteJSON(w, http.StatusOK, registeredUser); err != nil {
		return internalErrorResponseV2("an error occured", err)
	}

	return nil
}

func (a *application) handleLoginUser(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var user User
	if err := ReadJSON(w, r, &user); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	validate := validator.New()
	err := validate.Struct(user)
	if err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	registeredUser, err := a.storer.GetUserByEmail(ctx, user.Email)
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

	SuccessfulLoginResponse(w, token)
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

	resp, err := a.storer.GetTodosByUserID(ctx, userID, limit, page)
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

func (a *application) handleGetTodo(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid todo list id", err)
	}

	resp, err := a.storer.GetTodo(ctx, id)
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

func (a *application) handleCreateTodo(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var todo CreateTodo

	if err := ReadJSON(w, r, &todo); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	out, err := a.storer.CreateTodo(ctx, todo)
	if err != nil {
		return internalErrorResponseV2("failed to create todo list", err)
	}

	if err := WriteJSON(w, http.StatusOK, out); err != nil {
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

	if err := ReadJSON(w, r, &update); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	err = a.storer.UpdateTodo(ctx, id, update)
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

func (a *application) handleDeleteTodo(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid todo list id", err)
	}

	if err := a.storer.DeleteTodo(ctx, id); err != nil {
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

func (a *application) handleGetTodoItems(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	todoID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return badRequestResponseV2("invalid todo list id", err)
	}

	items, err := a.storer.GetTodoItems(ctx, todoID)
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

func (a *application) handleAddTodoItem(w http.ResponseWriter, r *http.Request) *apiErrorV2 {
	ctx := r.Context()

	var item Item

	if err := ReadJSON(w, r, &item); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	if err := a.storer.AddTodoItem(ctx, item); err != nil {
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

	if err := ReadJSON(w, r, &item); err != nil {
		return badRequestResponseV2("invalid data", err)
	}

	if err := a.storer.UpdateTodoItem(ctx, item); err != nil {
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

	if err := a.storer.DeleteTodoItem(ctx, id); err != nil {
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
