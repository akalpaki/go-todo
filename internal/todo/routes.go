package todo

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/akalpaki/todo/pkg/web"
)

func Routes(logger *slog.Logger, repository *Repository) http.Handler {
	mux := http.NewServeMux()

	// TODO routes
	mux.HandleFunc("POST /", web.Access(web.Auth(HandleCreate(logger, repository)), logger))
	mux.HandleFunc("GET /", web.Access(web.Auth(HandleGetForUser(logger, repository)), logger))
	mux.HandleFunc("GET /{id}", web.Access(web.Auth(HandleGetByID(logger, repository)), logger))
	mux.HandleFunc("PUT /{id}", web.Access(web.Auth(HandleUpdate(logger, repository)), logger))
	mux.HandleFunc("DELETE /{id}", web.Access(web.Auth(HandleDelete(logger, repository)), logger))

	// TASK routes
	mux.HandleFunc("POST /{id}/items", web.Access(web.Auth(HandleCreateTask(logger, repository)), logger))
	mux.HandleFunc("GET /{id}/items", web.Access(web.Auth(HandleGetTasks(logger, repository)), logger))
	mux.HandleFunc("PUT /{todo_id}/items/{task_id}", web.Access(web.Auth(HandleUpdateTask(logger, repository)), logger))
	mux.HandleFunc("DELETE /{todo_id}/items/{task_id}", web.Access(web.Auth(HandleDeleteTask(logger, repository)), logger))

	return mux
}

func HandleCreate(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		data, err := web.ReadJSON[TodoRequest](r)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid data or malformed json", err)
			return
		}

		authorID, ok := ctx.Value(web.UserID).(string)
		if !ok || data.AuthorID != authorID {
			web.ErrorResponse(logger, w, r, http.StatusForbidden, "you do not have access to this resource", web.ErrInvalidUserID)
			return
		}

		todo, err := repository.Create(ctx, data)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to create todo", err)
			return
		}

		if err := web.WriteJSON(w, r, http.StatusCreated, todo); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to produce response", err)
			return
		}
	}
}

func HandleGetForUser(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	const defaultLimit = 10
	const defaultPage = 1

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		queryParams := r.URL.Query()
		page, err := strconv.Atoi(queryParams.Get("page"))
		if err != nil {
			page = defaultPage
		}
		limit, err := strconv.Atoi(queryParams.Get("limit"))
		if err != nil {
			limit = defaultLimit
		}

		userID, ok := ctx.Value(web.UserID).(string)
		if !ok {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid user id", web.ErrInvalidUserID)
			return
		}

		todos, err := repository.GetByUserID(ctx, userID, limit, page)
		if err != nil {
			switch err {
			case errNoTodosForUser:
				web.ErrorResponse(logger, w, r, http.StatusNotFound, "no todos found for user", err)
				return
			default:
				web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to retrieve todo lists", err)
				return
			}
		}

		if err := web.WriteJSON(w, r, http.StatusOK, todos); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to produce response", err)
			return
		}
	}
}

func HandleGetByID(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("id")
		userID, ok := ctx.Value(web.UserID).(string)
		if !ok {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid user id", web.ErrInvalidUserID)
			return
		}

		todo, err := repository.GetByID(ctx, id)
		if err != nil {
			switch err {
			case errNotFound:
				web.ErrorResponse(logger, w, r, http.StatusNotFound, "todo not found", err)
				return
			default:
				web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to retrieve todo", err)
				return
			}
		}

		if userID != todo.AuthorID {
			web.ErrorResponse(logger, w, r, http.StatusForbidden, "you do not have access to this resource", err)
			return
		}

		if err := web.WriteJSON(w, r, http.StatusOK, todo); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to produce response", err)
			return
		}
	}
}

func HandleUpdate(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		todoID := r.PathValue("id")
		userID, ok := ctx.Value(web.UserID).(string)
		if !ok {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid user id", web.ErrInvalidUserID)
			return
		}

		todo, err := repository.GetByID(ctx, todoID)
		if err != nil {
			switch err {
			case errNotFound:
				web.ErrorResponse(logger, w, r, http.StatusNotFound, "todo not found", err)
				return
			default:
				web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to retrieve todo", err)
				return
			}
		}

		if userID != todo.AuthorID {
			web.ErrorResponse(logger, w, r, http.StatusForbidden, "you do not have access to this resource", err)
			return
		}

		update, err := web.ReadJSON[TodoRequest](r)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid data or malformed json", err)
			return
		}

		if err := repository.Update(ctx, todoID, update); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to update todo", err)
			return
		}

		if err := web.WriteJSON(w, r, http.StatusOK, ""); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to produce repsonse", err)
			return
		}
	}
}

func HandleDelete(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		todoID := r.PathValue("id")
		userID, ok := ctx.Value(web.UserID).(string)
		if !ok {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid user id", web.ErrInvalidUserID)
			return
		}

		todo, err := repository.GetByID(ctx, todoID)
		if err != nil {
			switch err {
			case errNotFound:
				web.ErrorResponse(logger, w, r, http.StatusNotFound, "todo not found", err)
				return
			default:
				web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to retrieve todo", err)
				return
			}
		}

		if userID != todo.AuthorID {
			web.ErrorResponse(logger, w, r, http.StatusForbidden, "you do not have access to this resource", err)
			return
		}

		if err := repository.DeleteTodo(ctx, todoID); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to delete todo", err)
			return
		}
	}
}

func HandleCreateTask(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		task, err := web.ReadJSON[Task](r)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid data or malformed json", err)
			return
		}

		if err := repository.CreateTask(ctx, task); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to create task", err)
			return
		}

		if err := web.WriteJSON(w, r, http.StatusOK, ""); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to produce response", err)
			return
		}
	}
}

func HandleGetTasks(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		todoID := r.PathValue("id")

		tasks, err := repository.GetTasks(ctx, todoID)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to retrieve tasks", err)
			return
		}

		if err := web.WriteJSON(w, r, http.StatusOK, tasks); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to produce response", err)
			return
		}
	}
}

func HandleUpdateTask(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		update, err := web.ReadJSON[Task](r)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid data or malformed json", err)
			return
		}

		if err := repository.UpdateTask(ctx, update); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to update task", err)
			return
		}

		if err := web.WriteJSON(w, r, http.StatusOK, ""); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to produce repsonse", err)
			return
		}
	}
}

func HandleDeleteTask(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id := r.PathValue("task_id")

		if err := repository.DeleteTask(ctx, id); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to delete task", err)
			return
		}

		if err := web.WriteJSON(w, r, http.StatusOK, ""); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "produce response", err)
			return
		}
	}
}
