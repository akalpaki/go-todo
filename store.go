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
		if err := rows.Scan(&td.ID, &td.Name); err != nil {
			return nil, err
		}
		res = append(res, td)
	}
	return res, nil
}

// GetTodo returns a specific todo list along with its items.
func (s *Storer) GetTodo(ctx context.Context, id int) (Todo, error) {
	var res Todo

	tdRow := s.DB.QueryRowContext(ctx, "select id, name from todo where id = ?", id)
	if err := tdRow.Scan(&res.ID, &res.Name); err != nil {
		return Todo{}, err
	}

	itemRows, err := s.DB.QueryContext(ctx, "select * from todo_item where todo_id = ?", id)
	if err != nil {
		return Todo{}, err
	}

	for itemRows.Next() {
		var item Item
		if err := itemRows.Scan(&item.ItemNo, &item.Content, &item.Done, &item.TodoID); err != nil {
			return Todo{}, err
		}
		res.Items = append(res.Items, item)
	}

	return res, nil
}

// CreateTodo creates a new Todo list. If todo items are passed, they are added to the list.
func (s *Storer) CreateTodo(ctx context.Context, ct CreateTodo) (Todo, error) {
	newTodo := Todo{
		Name:  ct.Name,
		Items: ct.Items,
	}

	addItem, prepErr := s.DB.Prepare("insert into todo_item(itemNo, content, done, todo_id) values (?,?,?,?)") // TODO: look into preparing these once somewhere, maybe sync once?
	if prepErr != nil {
		return Todo{}, prepErr
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return Todo{}, err
	}
	txAddItem := tx.Stmt(addItem)

	res, err := tx.Exec("insert into todo (name) values (?)", newTodo.Name)
	if err != nil {
		tx.Rollback()
		return Todo{}, err
	}
	todoID, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return Todo{}, err
	}
	newTodo.ID = int(todoID)

	if len(newTodo.Items) > 0 {
		for i, v := range newTodo.Items {
			_, err := txAddItem.Exec(v.ItemNo, v.Content, v.Done, todoID)
			if err != nil {
				tx.Rollback()
				return Todo{}, err
			}
			newTodo.Items[i].TodoID = int(todoID)
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return Todo{}, err
	}

	return newTodo, nil
}

// UpdateTodo updates a todo list's information
func (s *Storer) UpdateTodo(ctx context.Context, id int, updated UpdateTodo) error {
	if updated.Name != nil {
		_, err := s.DB.ExecContext(ctx, "update todo set name = ? where id = ?", updated.Name, id)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteTodo deletes an entire todo list.
func (s *Storer) DeleteTodo(ctx context.Context, id int) error {
	_, err := s.DB.Exec("delete from todo where id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

// GetTodoItems retrieves todo list items for a given todo list id
func (s *Storer) GetTodoItems(ctx context.Context, todoID int) ([]Item, error) {
	items := make([]Item, 0)

	rows, err := s.DB.QueryContext(ctx, "select * from todo_item where todo_id = ?", todoID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item Item
		if err := rows.Scan(&item); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Storer) AddTodoItem(ctx context.Context, item Item) error {
	_, err := s.DB.ExecContext(ctx, "insert into todo_item (itemNo, content, done, todo_id) values (?,?,?,?)", item.ItemNo, item.Content, item.Done, item.TodoID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Storer) UpdateTodoItem(ctx context.Context, update Item) error {
	_, err := s.DB.ExecContext(ctx, "update todo_item set content = ?, done = ? where todo_id = ? and itemNo = ?", update.Content, update.Done, update.TodoID, update.ItemNo)
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
