package response

type Login struct {
	Token          string
	CurrentAccount *AccountOne
}
