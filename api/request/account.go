package request

type AccountCreateOne struct {
	Name string `json:"name" binding:"required"`
}
