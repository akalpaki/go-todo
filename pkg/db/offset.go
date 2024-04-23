package db

func CalculateOffset(page, pageSize int) int {
	offset := (page - 1) * pageSize
	if offset < 0 {
		offset = offset * -1
	}
	return offset
}
