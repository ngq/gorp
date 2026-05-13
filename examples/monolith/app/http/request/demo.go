package request

type CreateDemo struct {
	Name string `json:"name" binding:"required"`
}

type UpdateDemo struct {
	Name string `json:"name" binding:"required"`
}
