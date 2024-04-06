package main

import (
	"context"
	"database/sql"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	errNotFound = errors.New("entry not found")
)

type repository struct {
	DB *sql.DB
}

func newRepository(db *sql.DB) *repository {
	return &repository{
		DB: db,
	}
}

func (r *repository) CreateUser(ctx context.Context, user User) (User, error) {
	var err error

	user.Password, err = hashPassword(user.Password)
	if err != nil {
		return User{}, err
	}

	res, err := r.DB.ExecContext(ctx, "insert into user (email, password) values (?, ?)", user.Email, user.Password)
	if err != nil {
		return User{}, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return User{}, err
	}
	user.ID = int(id)

	return user, nil
}

func (r *repository) GetUserByEmail(ctx context.Context, email string) (User, error) {
	var registeredUser User

	res := r.DB.QueryRowContext(ctx, "select * from user where email = ?", email)
	if err := res.Scan(&registeredUser.ID, &registeredUser.Email, &registeredUser.Password); err != nil {
		return User{}, err
	}
	return registeredUser, nil
}

func (r *repository) GetUserByID(id int) (User, error) {
	var registeredUser User
	row := r.DB.QueryRow("select * from user where id = ?", id)
	if err := row.Scan(&registeredUser); err != nil {
		return User{}, err
	}
	return registeredUser, nil
}

// GetTodosByUserID retrieves the previews of the todo lists for a specific user from the database. It does NOT return the items.
func (r *repository) GetTodosByUserID(ctx context.Context, userID int, limit, page int) ([]Todo, error) {
	res := make([]Todo, 0)

	offset := calculateOffset(page, limit)
	rows, err := r.DB.QueryContext(ctx, "select * from todo where user_id = ? order by id limit ? offset ?", userID, limit, offset)
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
func (r *repository) GetTodo(ctx context.Context, id int) (Todo, error) {
	var res Todo

	tdRow := r.DB.QueryRowContext(ctx, "select id, name, user_id from todo where id = ?", id)
	if err := tdRow.Scan(&res.ID, &res.Name, &res.UserID); err != nil {
		return Todo{}, err
	}

	itemRows, err := r.DB.QueryContext(ctx, "select * from todo_item where todo_id = ?", id)
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

// GetTodoMetadataByID returns only the metadata about a todo list stored in the todo table. Currently used for auth purposes.
func (r *repository) GetTodoMetadataByID(id int) (Todo, error) {
	var res Todo
	row := r.DB.QueryRow("select * from todo where id = ?", id)
	if err := row.Scan(&res); err != nil {
		return Todo{}, err
	}
	return res, nil
}

// CreateTodo creates a new Todo list. If todo items are passed, they are added to the list.
func (r *repository) CreateTodo(ctx context.Context, ct CreateTodo) (Todo, error) {
	newTodo := Todo{
		Name:   ct.Name,
		Items:  ct.Items,
		UserID: ct.UserID,
	}

	addItem, prepErr := r.DB.Prepare("insert into todo_item(itemNo, content, done, todo_id) values (?,?,?,?)") // TODO: look into preparing these once somewhere, maybe sync once?
	if prepErr != nil {
		return Todo{}, prepErr
	}

	tx, err := r.DB.BeginTx(ctx, nil)
	if err != nil {
		return Todo{}, err
	}
	txAddItem := tx.Stmt(addItem)

	res, err := tx.Exec("insert into todo (name, user_id) values (?, ?)", newTodo.Name, newTodo.UserID)
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
func (r *repository) UpdateTodo(ctx context.Context, id int, updated UpdateTodo) error {
	if updated.Name != nil {
		_, err := r.DB.ExecContext(ctx, "update todo set name = ? where id = ?", updated.Name, id)
		if err != nil {
			return err
		}
	}
	return nil
}

// DeleteTodo deletes an entire todo list.
func (r *repository) DeleteTodo(ctx context.Context, id int) error {
	_, err := r.DB.Exec("delete from todo where id = ?", id)
	if err != nil {
		return err
	}
	return nil
}

// GetTodoItems retrieves todo list items for a given todo list id
func (r *repository) GetTodoItems(ctx context.Context, todoID int) ([]Item, error) {
	items := make([]Item, 0)

	rows, err := r.DB.QueryContext(ctx, "select * from todo_item where todo_id = ?", todoID)
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

func (r *repository) AddTodoItem(ctx context.Context, item Item) error {
	_, err := r.DB.ExecContext(ctx, "insert into todo_item (itemNo, content, done, todo_id) values (?,?,?,?)", item.ItemNo, item.Content, item.Done, item.TodoID)
	if err != nil {
		return err
	}
	return nil
}

func (r *repository) UpdateTodoItem(ctx context.Context, update Item) error {
	_, err := r.DB.ExecContext(ctx, "update todo_item set content = ?, done = ? where todo_id = ? and itemNo = ?", update.Content, update.Done, update.TodoID, update.ItemNo)
	if err != nil {
		return err
	}
	return nil
}

// DeleteTodoItem deletes a specific todo item
func (r *repository) DeleteTodoItem(ctx context.Context, itemID int) error {
	_, err := r.DB.Exec("delete from todo_item where id = ?")
	if err != nil {
		return err
	}
	return nil
}

func calculateOffset(page, limit int) int {
	return (page - 1) * limit
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
