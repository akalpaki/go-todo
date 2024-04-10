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
	mux.HandleFunc("POST /", web.Auth(handleCreate(logger, repository)))
	mux.HandleFunc("GET /", web.Auth(handleGetForUser(logger, repository)))
	mux.HandleFunc("GET /{id}", web.Auth(handleGetByID(logger, repository)))
	mux.HandleFunc("PUT /{id}", web.Auth(handleUpdate(logger, repository)))
	mux.HandleFunc("DELETE /{id}", web.Auth(handleDelete(logger, repository)))

	// TASK routes
	mux.HandleFunc("POST /{id}/items", web.Auth(handleCreateTask(logger, repository)))
	mux.HandleFunc("GET /{id}/items", web.Auth(handleGetTasks(logger, repository)))
	mux.HandleFunc("PUT /{todo_id}/items/{task_id}", web.Auth(handleUpdateTask(logger, repository)))
	mux.HandleFunc("DELETE /{todo_id}/items/{task_id}", web.Auth(handleDeleteTask(logger, repository)))

	return mux
}

func handleCreate(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		data, err := web.ReadJSON[TodoRequest](r)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusBadRequest, "invalid data or malformed json", err)
			return
		}

		todo, err := repository.Create(ctx, data)
		if err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to create todo", err)
			return
		}

		if err := web.WriteJSON(w, r, http.StatusOK, todo); err != nil {
			web.ErrorResponse(logger, w, r, http.StatusInternalServerError, "failed to produce response", err)
			return
		}
	}
}

func handleGetForUser(logger *slog.Logger, repository *Repository) http.HandlerFunc {
	const defaultLimit = 10

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		queryParams := r.URL.Query()
		page, err := strconv.Atoi(queryParams.Get("page"))
		if err != nil {
			page = 0
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

func handleGetByID(logger *slog.Logger, repository *Repository) http.HandlerFunc {
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

func handleUpdate(logger *slog.Logger, repository *Repository) http.HandlerFunc {
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

func handleDelete(logger *slog.Logger, repository *Repository) http.HandlerFunc {
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

func handleCreateTask(logger *slog.Logger, repository *Repository) http.HandlerFunc {
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

func handleGetTasks(logger *slog.Logger, repository *Repository) http.HandlerFunc {
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

func handleUpdateTask(logger *slog.Logger, repository *Repository) http.HandlerFunc {
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

func handleDeleteTask(logger *slog.Logger, repository *Repository) http.HandlerFunc {
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
