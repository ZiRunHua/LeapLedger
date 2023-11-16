package initialize

type _thirdParty struct {
	WeCom _weCom `yaml:"WeCom"`
}

type _weCom struct {
	CorpId     string `yaml:"CorpId"`
	CorpSecret string `yaml:"CorpSecret"`
}
