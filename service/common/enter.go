package commonService

type Group struct {
	common
	current
}

var GroupApp = new(Group)
