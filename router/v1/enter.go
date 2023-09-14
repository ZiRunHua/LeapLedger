package v1

type RouterGroup struct {
	AccountRouter
	CategoryRouter
	UserRouter
	PublicRouter
	TransactionImportRouter
}
