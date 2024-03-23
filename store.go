package main

import (
	"context"
	"database/sql"
)

type Storer struct {
	DB *sql.DB
}

func NewStorer(db *sql.DB) *Storer {
	return &Storer{
		DB: db,
	}
}

// GetTodos retrieves the previews of the todo lists from the database. It does NOT return the items.
func (s *Storer) GetTodos(ctx context.Context) ([]Todo, error) {
	res := make([]Todo, 0)

	// TODO: pagination
	rows, err := s.DB.QueryContext(ctx, "select * from todo")
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var td Todo
		if err := rows.Scan(&td); err != nil {
			return nil, err
		}
		res = append(res, td)
	}
	return res, nil
}

// GetTodo returns a specific todo list along with its items.
func (s *Storer) GetTodo(ctx context.Context, id int) (Todo, error) {
	var res Todo

	tdRow := s.DB.QueryRowContext(ctx, "select * from todo where id = ?", id)
	if err := tdRow.Scan(&res); err != nil {
		return Todo{}, err
	}

	itemRows, err := s.DB.QueryContext(ctx, "select * from todo_item where todo_id = ?", id)
	if err != nil {
		return Todo{}, err
	}

	for itemRows.Next() {
		var item TodoItem
		if err := itemRows.Scan(&item); err != nil {
			return Todo{}, err
		}
		res.Items = append(res.Items, item)
	}

	return res, nil
}

// CreateTodo creates a new Todo list. If todo items are passed, they are added to the list.
func (s *Storer) CreateTodo(ctx context.Context, td Todo) (Todo, error) {
	addItem, prepErr := s.DB.Prepare("insert into todo_item(content, done, todo_id) values (?, ?, ?)") // TODO: look into preparing these once somewhere, maybe sync once?
	if prepErr != nil {
		return Todo{}, prepErr
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return Todo{}, err
	}
	txAddItem := tx.Stmt(addItem)

	_, err = tx.Exec("insert into todo (name) values ?", td.Name)
	if err != nil {
		tx.Rollback()
		return Todo{}, err
	}

	if len(td.Items) > 0 {
		for i, v := range td.Items {
			res, err := txAddItem.Exec(v.Content, v.Done, v.TodoID)
			if err != nil {
				tx.Rollback()
				return Todo{}, err
			}
			id, err := res.LastInsertId()
			if err != nil {
				tx.Rollback()
				return Todo{}, err
			}
			td.Items[i].ID = int(id)
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return Todo{}, err
	}

	return td, nil
}

func (s *Storer) UpdateTodo(ctx context.Context, id int, updated Todo) (Todo, error) {
	updateItem, prepErr := s.DB.Prepare("update table todo_item set content = ?, done = ? where id = ?")
	if prepErr != nil {
		return Todo{}, prepErr
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return Todo{}, err
	}

	txUpdateItem := tx.Stmt(updateItem)

	_, err = tx.Exec("update table todo name = ? where id = ?", updated.Name, id)
	if err != nil {
		tx.Rollback()
		return Todo{}, err
	}

	for _, v := range updated.Items {
		_, err := txUpdateItem.Exec(v.Content, v.Done, v.ID)
		if err != nil {
			tx.Rollback()
			return Todo{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return Todo{}, err
	}

	return updated, nil
}

// DeleteTodo deletes an entire todo list.
func (s *Storer) DeleteTodo(ctx context.Context, id int) error {
	_, err := s.DB.Exec("delete from todo where id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTodoItem deletes a specific todo item
func (s *Storer) DeleteTodoItem(ctx context.Context, itemID int) error {
	_, err := s.DB.Exec("delete from todo_item where id = ?")
	if err != nil {
		return err
	}
	return nil
}
