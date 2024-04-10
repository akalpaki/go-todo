package todo

// Todo is the model that represents the Todo list entity.
type Todo struct {
	ID       string
	AuthorID string
	Name     string
	Tasks    []Task
}

// TodoRequest is the model containing the minimum required information to create and update a todo list.
// Requests should always be validated with the Valid methods before being accepted.
type TodoRequest struct {
	AuthorID string
	Name     string
	Tasks    []Task
}

func (r TodoRequest) Valid() bool {
	return r.AuthorID != "" && r.Name != "" && r.Tasks != nil
}

// Task is the model that represents a single Todo list task.
type Task struct {
	ID      string
	TodoID  string
	Content string
	Done    bool
	Order   int
}

func (r Task) Valid() bool {
	return r.TodoID != "" && r.Content != "" && r.Order > 0
}
