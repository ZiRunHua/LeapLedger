package commonModel

type Model interface {
	modelInterface()
}

type BaseModel struct {
}

func (base *BaseModel) modelInterface() {

}
