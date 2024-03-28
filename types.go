package main

// User is the model the represent an individual user
type User struct {
	ID       int    `json:"user_id"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// Todo is the model that represents a todo list entity.
type Todo struct {
	ID     int    `json:"id"`
	UserID int    `json:"user_id"`
	Name   string `json:"name"`
	Items  []Item `json:"items"`
}

// CreateTodo is the model that represents the needed information to create a new todo list.
type CreateTodo struct {
	Name  string `json:"name"`
	Items []Item `json:"items"`
}

// UpdateTodo is the model that represents any potential information that might be updated on a
// todo list. The fields are pointers as to allow for nil values, denoting them as optional.
type UpdateTodo struct {
	Name *string `json:"name"`
}

// Item is the model for the value object representing a single item in a todo list.
type Item struct {
	ItemNo  int    `json:"itemNo"`
	Content string `json:"content"`
	Done    bool   `json:"done"`
	TodoID  int    `json:"todo_id"`
}
