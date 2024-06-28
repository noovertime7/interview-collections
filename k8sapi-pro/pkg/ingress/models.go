package ingress

type IngressModel struct {
	Name       string
	Namespace  string
	CreateTime string
	Host       string
	Options    IngressOptions
}

type IngressOptions struct {
	IsCros    bool
	IsCanary  bool
	IsRewrite bool
}

type IngressPost struct {
	Name        string
	Namespace   string
	Annotations string
	Rules       []*IngressRules
}

type IngressRules struct {
	Host  string         `json:"host"`
	Paths []*IngressPath `json:"paths"`
}

type IngressPath struct {
	Path    string `json:"path"`
	SvcName string `json:"svc_name"`
	Port    string `json:"port"`
}
