package todo

// Todo is the model that represents the Todo list entity.
type Todo struct {
	ID       string `json:"id"`
	AuthorID string `json:"author_id"`
	Name     string `json:"name"`
	Tasks    []Task `json:"tasks"`
}

// TodoRequest is the model containing the minimum required information to create and update a todo list.
// Requests should always be validated with the Valid methods before being accepted.
type TodoRequest struct {
	AuthorID string `json:"author_id"`
	Name     string `json:"name"`
	Tasks    []Task `json:"tasks"`
}

func (r TodoRequest) Valid() bool {
	return r.AuthorID != "" && r.Name != "" && r.Tasks != nil
}

// Task is the model that represents a single Todo list task.
type Task struct {
	ID      string `json:"task_id"`
	TodoID  string `json:"todo_id"`
	Content string `json:"content"`
	Done    bool   `json:"done"`
	Order   int    `json:"order"`
}

func (r Task) Valid() bool {
	return r.Content != "" && r.Order >= 0
}
