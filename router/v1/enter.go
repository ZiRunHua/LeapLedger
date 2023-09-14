package v1

type RouterGroup struct {
	AccountRouter
	CategoryRouter
	TransactionRouter
	UserRouter
	PublicRouter
	TransactionImportRouter
}
