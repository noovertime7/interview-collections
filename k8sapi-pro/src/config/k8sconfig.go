package config

import (
	"gopkg.in/yaml.v2"
	"gorm.io/gorm"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/restmapper"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/client-go/util/workqueue"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"k8sapi-pro/pkg/namespace"
	"k8sapi-pro/src/caches"
	"k8sapi-pro/src/models"
	"log"
	"os"
	"strings"
	"time"
)

const (
	RESOURCE = "./resources"
	PATH     = "./resources/config"
)

type K8sConfig struct {
	RHandler  *caches.ResourceHandler `inject:"-"`
	NsHandler *namespace.NsHandler    `inject:"-"`
	Db        *gorm.DB                `inject:"-"` //测试限速队列消费者操作数据库
}

func NewK8sConfig() *K8sConfig {
	return &K8sConfig{}
}

func (this *K8sConfig) GetRestConfigMap() *models.RestConfigMap {
	set := models.RestConfigMap{}
	files, err := ioutil.ReadDir(RESOURCE)
	if err != nil {
		log.Fatal(err)
	}
	for _, fileInfo := range files {
		file, err := os.Open(strings.Join([]string{RESOURCE, fileInfo.Name()}, "/"))
		if err != nil {
			log.Fatal(err)
		}
		b, err := ioutil.ReadAll(file)
		if err != nil {
			log.Fatal(err)
		}
		cliConf, err := clientcmd.NewClientConfigFromBytes(b)
		if err != nil {
			log.Fatal(err)
		}
		apiConf, err := cliConf.RawConfig()
		if err != nil {
			log.Fatal(err)
		}
		cluster := apiConf.Contexts[apiConf.CurrentContext].Cluster
		//restCli, err := cliConf.ClientConfig()
		restCli, err := clientcmd.BuildConfigFromKubeconfigGetter("", func() (*api.Config, error) {
			return &apiConf, nil
		})
		if err != nil {
			log.Fatal(err)
		}
		set[cluster] = restCli
	}
	return &set
}

func (this *K8sConfig) InitClientSet() *models.CliMap {
	set := models.CliMap{}
	for cluster, restConf := range *this.GetRestConfigMap() {
		if cli, err := kubernetes.NewForConfig(restConf); err == nil {
			set[cluster] = cli
		} else {
			log.Fatal(err)
		}
	}
	return &set
}

func (this *K8sConfig) InitDynamicClient() *models.DcliMap {
	set := models.DcliMap{}
	for cluster, restConf := range *this.GetRestConfigMap() {
		if dc, err := dynamic.NewForConfig(restConf); err == nil {
			set[cluster] = dc
		} else {
			log.Fatal(err)
		}
	}
	return &set
}

// metrics客户端
func (this *K8sConfig) InitMetricClient() *models.McliMap {
	set := models.McliMap{}
	for cluster, restConf := range *this.GetRestConfigMap() {
		if mc, err := versioned.NewForConfig(restConf); err == nil {
			set[cluster] = mc
		} else {
			log.Fatal(err)
		}
	}
	return &set
}

// 脚手架通过config()对里面的opt进行依赖注入，强制所有方法都要返回指针类型（即使不注入）
func (this *K8sConfig) ParseGVR() *models.GVRs {
	// 读取 YAML 文件
	yamlData, err := ioutil.ReadFile("gvr.yaml")
	if err != nil {
		panic(err)
	}

	// 解析 YAML
	var gvrInfoList models.GVRs
	err = yaml.Unmarshal(yamlData, &gvrInfoList)
	if err != nil {
		panic(err)
	}

	return &gvrInfoList
}

func (this *K8sConfig) InitDynamicInformerFactory() *models.InformerList {
	set := models.InformerList{}
	for cluster, dc := range *this.InitDynamicClient() {
		fac := dynamicinformer.NewDynamicSharedInformerFactory(dc, time.Minute*5)
		//这里面临的问题是，因为handler的成员变量使用了依赖注入，所以无法通过构造函数来生成不同的handler
		//又因为handler本身也是依赖注入的，如何为不同集群的informer注册带有集群标示的handler就成了问题
		//因此使用了装饰器模式，包装了一个注册用的handler，以原handler为基底，这样就可以实现参数传入
		h := caches.WrapperHandler(this.RHandler, cluster)

		for _, gvr := range *this.ParseGVR() {
			_, err := fac.ForResource(gvr).Informer().AddEventHandler(h)
			if err != nil {
				log.Fatal(err)
			}
		}

		//nsh := namespace.WrapperNsHandler(this.NsHandler, cluster)
		//_, err = fac.ForResource(schema.GroupVersionResource{Group: "", Version: "v1", Resource: "namespaces"}).
		//	Informer().AddEventHandler(nsh)
		//if err != nil {
		//	log.Fatal(err)
		//}

		fac.Start(wait.NeverStop)
		set = append(set, fac)
	}
	return &set
}

//todo single
//func (this *K8sConfig) InitDynamicCli() *dynamic.DynamicClient {
//	rc, err := clientcmd.BuildConfigFromFlags("", PATH)
//	if err != nil {
//		log.Fatal(err)
//	}
//	if dc, err := dynamic.NewForConfig(rc); err == nil {
//		return dc
//	} else {
//		log.Fatal(err)
//	}
//	return nil
//}
//
//func (this *K8sConfig) InitDynamicInformerFactory() dynamicinformer.DynamicSharedInformerFactory {
//	fac := dynamicinformer.NewDynamicSharedInformerFactory(this.InitDynamicCli(), 0)
//
//	_, err := fac.ForResource(schema.GroupVersionResource{Group: "apps", Version: "v1", Resource: "deployments"}).
//		Informer().AddEventHandler(this.RHandler)
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	fac.Start(wait.NeverStop)
//	return fac
//}

func (this *K8sConfig) InitK8sCli() *kubernetes.Clientset {
	rc, err := clientcmd.BuildConfigFromFlags("", PATH)
	if err != nil {
		log.Fatal(err)
	}
	cs, err := kubernetes.NewForConfig(rc)
	if err != nil {
		log.Fatal(err)
	}
	return cs
}

func (this *K8sConfig) GetRestMapper() *meta.RESTMapper {
	//用generic的configFlag可以直接获取
	apigr, err := restmapper.GetAPIGroupResources(this.InitK8sCli().Discovery())
	if err != nil {
		log.Fatal(err)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(apigr)
	return &mapper
}

// 初始化 系统 配置
func (*K8sConfig) InitSysConfig() *models.SysConfig {
	b, err := ioutil.ReadFile("node.yaml")
	if err != nil {
		log.Fatal(err)
	}
	config := &models.SysConfig{}
	err = yaml.Unmarshal(b, config)
	if err != nil {
		log.Fatal(err)
	}
	return config
}

func (*K8sConfig) InitQ() *workqueue.RateLimitingInterface {
	// 创建工作队列
	queue := workqueue.NewNamedRateLimitingQueue(workqueue.DefaultControllerRateLimiter(), "rateLimit")
	return &queue
}
