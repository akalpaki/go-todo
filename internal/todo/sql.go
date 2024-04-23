package todo

const (
	selectTodoQuery            = "SELECT id, name, author_id FROM todos WHERE id = $1"
	selectTaskByTaskIDQuery    = "SELECT id, todo_id, task_order, content, done FROM tasks WHERE id = $1"
	selectTaskByTodoIDQuery    = "SELECT id, todo_id, task_order, content, done FROM tasks WHERE todo_id = $1"
	selectTodosByAuthorIDQuery = "SELECT id, author_id, name FROM todos WHERE author_id = $1 ORDER BY id LIMIT $2 OFFSET $3"
	insertTodoQuery            = "INSERT INTO todos (id, author_id, name) VALUES ($1, $2, $3)"
	insertTaskQuery            = "INSERT INTO tasks (id, todo_id, task_order, content, done) VALUES ($1, $2, $3, $4, $5)"
	updateTodoQuery            = "UPDATE todos SET name = $1 WHERE id = $2"
	updateTaskQuery            = "UPDATE tasks SET content = $1, done = $2 WHERE id = $3"
	deleteTodoQuery            = "DELETE FROM todos WHERE id = $1"
	deleteTaskQuery            = "DELETE FROM tasks WHERE id = $1"
)
