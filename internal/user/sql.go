package user

const (
	insert       = "INSERT INTO users (id, email, password) VALUES ($1, $2, $3)"
	queryByEmail = "SELECT id, email, password FROM users WHERE email = $1"
	queryByID    = "SELECT id, email, password FROM users WHERE id = $1"
)
