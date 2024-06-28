package models

type CertModel struct {
	CN        string //域名
	Algorithm string //算法
	Issuer    string //签发者
	BeginTime string //生效时间
	EndTime   string //到期时间
}
