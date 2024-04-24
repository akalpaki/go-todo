package todo

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/noquark/nanoid"
)

var (
	errNotFound       = errors.New("not found")
	errNoTodosForUser = errors.New("no todos found for user")
)

type Repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{
		pool: pool,
	}
}

//|+++++++++++++++++++++++++++++++++++++|
//|              TODO CRUD              |
//|+++++++++++++++++++++++++++++++++++++|

func (r *Repository) Create(ctx context.Context, data TodoRequest) (Todo, error) {
	conn, err := r.pool.Acquire(ctx)
	if err != nil {
		return Todo{}, fmt.Errorf("todo_repo acquire conn: %w", err)
	}
	defer conn.Release()

	id, err := nanoid.New(21)
	if err != nil {
		return Todo{}, nil
	}

	t := Todo{
		ID:       id,
		AuthorID: data.AuthorID,
		Name:     data.Name,
		Tasks:    data.Tasks,
	}

	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Todo{}, fmt.Errorf("todo_repo begin tx: %w", err)
	}

	_, err = tx.Exec(ctx, insertTodoQuery, t.ID, t.AuthorID, t.Name)
	if err != nil {
		tx.Rollback(ctx)
		return Todo{}, fmt.Errorf("todo_repo insert todo: %w", err)
	}

	if len(t.Tasks) > 0 {
		for _, v := range t.Tasks {
			_, err := tx.Exec(ctx, insertTaskQuery, v.ID, t.ID, v.Order, v.Content, v.Done) // note t.ID for the todo_id field because we've just generated it
			if err != nil {
				tx.Rollback(ctx)
				return Todo{}, fmt.Errorf("todo_repo insert task: %w", err)
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		tx.Rollback(ctx)
		return Todo{}, fmt.Errorf("todo_repo commit: %w", err)
	}

	return t, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Todo, error) {
	var t Todo

	tRow := r.pool.QueryRow(ctx, selectTodoQuery, id)
	if err := tRow.Scan(&t.ID, &t.AuthorID, &t.Name); err != nil {
		return Todo{}, fmt.Errorf("todo_repo get todo: %w", err)
	}

	iRows, err := r.pool.Query(ctx, selectTaskByTodoIDQuery, t.ID)
	if err != nil {
		return Todo{}, fmt.Errorf("todo_repo get tasks: %w", err)
	}

	tasks := make([]Task, 0)
	for iRows.Next() {
		var task Task
		if err := iRows.Scan(&task.ID, &task.TodoID, &task.Order, &task.Content, &task.Done); err != nil {
			return Todo{}, fmt.Errorf("todo_repo scan task: %w", err)
		}
		tasks = append(tasks, task)
	}
	t.Tasks = tasks

	return t, nil
}

func (r *Repository) GetByUserID(ctx context.Context, userID string, limit, page int) ([]Todo, error) {
	todos := make([]Todo, 0)

	offset := (page - 1) * limit
	rows, err := r.pool.Query(ctx, selectTodosByAuthorIDQuery, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("todo_repo select todos by userID: %w", err)
	}

	for rows.Next() {
		var t Todo
		if err := rows.Scan(&t.ID, &t.AuthorID, &t.Name); err != nil {
			return nil, fmt.Errorf("todo_repo scan todo: %w", err)
		}
		todos = append(todos, t)
	}

	if len(todos) == 0 {
		return nil, errNoTodosForUser
	}

	return todos, nil
}

func (r *Repository) Update(ctx context.Context, id string, update TodoRequest) error {
	if update.Name != "" {
		_, err := r.pool.Exec(ctx, updateTodoQuery, update.Name, id)
		if err != nil {
			return fmt.Errorf("todo_repo update todo: %w", err)
		}
	}
	return nil
}

func (r *Repository) DeleteTodo(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, deleteTodoQuery, id)
	if err != nil {
		return fmt.Errorf("todo_repo delete todo: %w", err)
	}

	return nil
}

//|++++++++++++++++++++++++++++++++|
//|           TASK CRUD            |
//|++++++++++++++++++++++++++++++++|

func (r *Repository) CreateTask(ctx context.Context, task Task) error {
	id, err := nanoid.New(0)
	if err != nil {
		return fmt.Errorf("todo_repo generating id: %w", err)
	}

	_, err = r.pool.Exec(ctx, insertTaskQuery, id, task.TodoID, task.Order, task.Content, task.Done)
	if err != nil {
		return fmt.Errorf("todo_repo insert task")
	}

	return nil
}

func (r *Repository) GetTasks(ctx context.Context, todoID string) ([]Task, error) {
	rows, err := r.pool.Query(ctx, selectTaskByTodoIDQuery, todoID)
	if err != nil {
		return nil, fmt.Errorf("todo_repo select tasks: %w", err)
	}

	tasks := make([]Task, 0)
	for rows.Next() {
		var task Task
		if err := rows.Scan(&task.ID, &task.TodoID, &task.Order, &task.Content, &task.Done); err != nil {
			return nil, fmt.Errorf("todo_repo scan task: %w", err)
		}
		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *Repository) UpdateTask(ctx context.Context, update Task) error {
	_, err := r.pool.Exec(ctx, updateTaskQuery, update.Content, update.Done, update.ID)
	if err != nil {
		return fmt.Errorf("todo_repo update task: %w", err)
	}
	return nil
}

func (r *Repository) DeleteTask(ctx context.Context, id string) error {
	_, err := r.pool.Exec(ctx, deleteTaskQuery, id)
	if err != nil {
		return fmt.Errorf("todo_repository delete task: %w", err)
	}

	return nil
}
