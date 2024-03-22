package main

import (
	"context"
	"database/sql"
)

const MIGRATION_TEMP = `
create table if not exists todo (id integer not null primary key autoincrement, name text not null);
create table if not exists todo_item (id integer not null primary key autoincrement, content text, done boolean not null, todo_id integer not null, foreign key (todo_id) references todo (id) on delete cascade);
`

type Storer struct {
	DB *sql.DB
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

	err = tx.Commit()
	if err != nil {
		return Todo{}, err
	}

	return td, nil
}

// AddTodoItems adds any number of todo items for a specific todo list
func (s *Storer) AddTodoItems(ctx context.Context, items []TodoItem) ([]TodoItem, error) {
	addItem, err := s.DB.Prepare("insert into todo_item (content, done, todo_id) values (?, ?, ?)")
	if err != nil {
		return nil, err
	}

	tx, err := s.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}

	txAddItem := tx.Stmt(addItem)
	for i, v := range items {
		res, err := txAddItem.ExecContext(ctx, v.Content, v.Done, v.TodoID)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		id, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		items[i].ID = int(id)
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return items, err
}

// UpdateTodoItem can be used to update an individual item from a todo list
func (s *Storer) UpdateTodoItem(ctx context.Context, item TodoItem) (TodoItem, error) {
	_, err := s.DB.ExecContext(ctx, "update todo_item (content, done) where id = ?", item.ID)
	if err != nil {
		return TodoItem{}, err
	}
	return item, err
}

func (s *Storer) DeleteTodo(ctx context.Context, id int) error {
	_, err := s.DB.ExecContext(ctx, "delete from todo where id = ?", id)
	return err
}

func (s *Storer) DeleteTodoItem(ctx context.Context, id int) error {
	_, err := s.DB.ExecContext(ctx, "delete from todo_item where id = ?", id)
	return err
}

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
