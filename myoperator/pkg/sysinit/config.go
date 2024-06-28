package sysinit

import (
	"github.com/gorilla/mux"
	"io/ioutil"
	"k8s.io/api/networking/v1"
	"os"
	"sigs.k8s.io/yaml"
)

type Server struct {
	Port int //代表是代理启动端口
}
type SysConfigStruct struct {
	Server  Server
	Ingress []v1.Ingress
}

var SysConfig = new(SysConfigStruct)

func InitConfig() error {
	config, err := ioutil.ReadFile("./app.yaml")
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(config, SysConfig)
	if err != nil {
		return err
	}
	ParseRule()
	return nil
}

// reconcile 有新的ingress被apply就修改内存中的配置，并持久化
// 还需要刷新router，类似于reload
// 方法可以直接在这里触发，也可以暴露外部api来触发。暴露了api就可以通过apply新的yaml触发，或者修改app.yaml再通过api来reload
func UpdateConfig(resource *v1.Ingress) error {
	isUpdate := false
	for i, ingress := range SysConfig.Ingress {
		if ingress.Name == resource.Name && ingress.Namespace == ingress.Namespace {
			SysConfig.Ingress[i] = *resource
			isUpdate = true
			break
		}
	}
	if !isUpdate {
		SysConfig.Ingress = append(SysConfig.Ingress, *resource)
	} // 写入内存

	err := save2File()
	if err != nil {
		return nil
	}

	return Reload() //重新加载路由
}

func DeleteIngress(name, namespace string) error {
	haveFound := false
	for i, ingress := range SysConfig.Ingress {
		if ingress.Name == name && ingress.Namespace == namespace {
			SysConfig.Ingress = append(SysConfig.Ingress[:i], SysConfig.Ingress[i+1:]...)
			haveFound = true
			break
		}
	}
	if haveFound {
		err := save2File()
		if err != nil {
			return err
		}
		return Reload()
	}
	return nil
}

func save2File() error {
	b, err := yaml.Marshal(SysConfig)
	if err != nil {
		return err
	}
	appYaml, err := os.OpenFile("./app.yaml", os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
	defer appYaml.Close()
	if err != nil {
		return err
	}
	_, err = appYaml.Write(b) // 持久化app.yaml
	if err != nil {
		return err
	}
	return nil
} //写入app.yaml文件

func Reload() error {
	MyRouter = mux.NewRouter() // warning 短暂的会有请求进来转发不了,应该为新router设置一个临时变量，然后在添加了路由后再赋值给MyRouter
	return InitConfig()
} //清空router的route，并重新从配置文件读取解析，生成route
