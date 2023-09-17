package request

type GetOne struct {
	Id int `json:"id" binding:"required"`
}
