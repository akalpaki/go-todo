package todo

const (
	selectTodoQuery            = "SELECT id, name, author_id FROM todo WHERE id = $1"
	selectTaskByTaskIDQuery    = "SELECT id, todo_id, order, content, done FROM task WHERE id = $1"
	selectTaskByTodoIDQuery    = "SELECT id, todo_id, order, content, done FROM task WHERE todo_id = $1"
	selectTodosByAuthorIDQuery = "SELECT id, name, author_id FROM todo WHERE author_id = $1 ORDER BY id LIMIT $2 OFFSET $3"
	insertTodoQuery            = "INSERT INTO todo (id, name, author_id) VALUES ($1, $2, $3)"
	insertTaskQuery            = "INSERT INTO task (id, todo_id, order, content, done) VALUES ($1, $2, $3, $4, $5)"
	updateTodoQuery            = "UPDATE todo SET name = $1 WHERE id = $2"
	updateTaskQuery            = "UPDATE task SET content = $1, done = $2 WHERE id = $3"
	deleteTodoQuery            = "DELETE FROM todo WHERE id = $1"
	deleteTaskQuery            = "DELETE FROM task WHERE id = $1"
)
