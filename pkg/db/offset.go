package db

func CalculateOffset(page, limit int) int {
	return (page - 1) * limit
}
