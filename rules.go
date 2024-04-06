package main

func userCanAccessTodo(userID int, td Todo) bool {
	return userID == td.UserID
}
