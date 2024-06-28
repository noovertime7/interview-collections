package configmap

type ConfigMapModel struct {
	Name       string
	NameSpace  string
	CreateTime string
	Content    map[string]string
}

type PostConfigMap struct {
	Name      string
	NameSpace string
	Data      map[string]string
	IsUpdate  bool //更新还是新建操作
}
