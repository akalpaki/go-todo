package user

const (
	insert       = "INSERT INTO user (id, email, password) VALUES ($1, $2, $3)"
	queryByEmail = "SELECT (id, email, password) FROM user WHERE email = $1"
	queryByID    = "SELECT (id, email, password) FROM user WHERE id = $1"
)
