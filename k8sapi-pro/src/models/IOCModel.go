package models

import (
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/metrics/pkg/client/clientset/versioned"
)

//防止循环引入，独立出package存放模型
//原位置在K8sConfig，原来的流向是————
//pkg/deployment/controller 引入 src/config/K8sConfig 用于创建时的apiserver调用
//src/config/K8sService 引入 src/deployment/service 用于依赖注入

type RestConfigMap map[string]*rest.Config

type CliMap map[string]*kubernetes.Clientset

type DcliMap map[string]*dynamic.DynamicClient

type McliMap map[string]*versioned.Clientset

type InformerList []dynamicinformer.DynamicSharedInformerFactory

type GVRs []schema.GroupVersionResource
