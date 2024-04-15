package initialize

type _thirdParty struct {
	WeCom _weCom `yaml:"WeCom"`
	Ai    _ai    `yaml:"Ai"`
}

type _weCom struct {
	CorpId     string `yaml:"CorpId"`
	CorpSecret string `yaml:"CorpSecret"`
}

type _ai struct {
	Host          string  `yaml:"Host"`
	Port          string  `yaml:"Port"`
	MinSimilarity float32 `yaml:"MinSimilarity"`
}

func (a _ai) GetPortalSite() string {
	return "http://" + a.Host + ":" + a.Port
}
