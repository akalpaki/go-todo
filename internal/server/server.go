package server

import (
	"github.com/akalpaki/todo/internal/todo"
	"github.com/akalpaki/todo/internal/user"
)

type Server struct {
	UserService user.Service
	TodoService todo.Service
}
