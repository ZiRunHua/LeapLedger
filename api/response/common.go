package response

type CommonCaptcha struct {
	CaptchaId     string `json:"captcha_id"`
	PicPath       string `json:"pic_path"`
	CaptchaLength int    `json:"captcha_length"`
	OpenCaptcha   bool   `json:"open_captcha"`
}
type Id struct {
	Id uint
}

type CreateResponse struct {
	Id        uint
	CreatedAt int64
	UpdatedAt int64
}

type Token struct {
	Token string
}
type TwoLevelTree struct {
	Tree []Father
}
type Father struct {
	NameId
	Children []NameId
}
type NameId struct {
	Id   uint
	Name string
}
type NameValue struct {
	Name  string
	Value int
}

type PageData struct {
	page  int
	limit int
	count int
}
