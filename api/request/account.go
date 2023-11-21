package request

type AccountCreateOne struct {
	Name string `binding:"required"`
	Icon string `binding:"required"`
}

type AccountUpdateOne struct {
	Name *string
	Icon *string
}
