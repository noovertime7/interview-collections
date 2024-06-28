package ingress

import (
	"context"
	"gorm.io/gorm"
	v1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8sapi-pro/src/models"
	"k8sapi-pro/src/utils"
	"strconv"
	"strings"
)

const (
	OPTIONS_CROS = iota
	OPTIONS_CANARY
	OPTIONS_REWRITE
)

const (
	ANNOTATIONS_CROS    = "nginx.ingress.kubernetes.io/enable-cors"
	ANNOTATIONS_CANARY  = "nginx.ingress.kubernetes.io/canary"
	ANNOTATIONS_REWRITE = "nginx.ingress.kubernetes.io/rewrite-enable"
)

type IngressSvc struct {
	Db       *gorm.DB       `inject:"-"`
	ClintMap *models.CliMap `inject:"-"`
}

func NewIngressSvc() *IngressSvc {
	return &IngressSvc{}
}

func (this *IngressSvc) ListAll(cluster, ns string) (ret []*IngressModel) {
	if list, err := models.List(cluster, ns, "Ingress", this.Db); err == nil {
		for _, resource := range list {
			obj := utils.Convert([]byte(resource.Object)).(*v1.Ingress)
			ret = append(ret, &IngressModel{
				Name:       obj.Name,
				Namespace:  obj.Namespace,
				CreateTime: obj.CreationTimestamp.Format("2006-01-02 15:04:05"),
				Host:       obj.Spec.Rules[0].Host,
				Options: IngressOptions{
					IsCros:    this.GetAnnotations(OPTIONS_CROS, obj),
					IsCanary:  this.GetAnnotations(OPTIONS_CANARY, obj),
					IsRewrite: this.GetAnnotations(OPTIONS_REWRITE, obj),
				},
			})
		}
	}
	return
}

// 获取列表时获取（固定的）annotations标签返回给前端显示用
// ingress的标签：如是否开启跨域，是否开启灰度
func (this *IngressSvc) GetAnnotations(options int, ingress *v1.Ingress) bool {
	switch options {
	case OPTIONS_CROS:
		if value, ok := ingress.ObjectMeta.Annotations[ANNOTATIONS_CROS]; ok {
			parseBool, _ := strconv.ParseBool(value)
			return parseBool
		}
		break
	case OPTIONS_CANARY:
		if value, ok := ingress.ObjectMeta.Annotations[ANNOTATIONS_CANARY]; ok {
			parseBool, _ := strconv.ParseBool(value)
			return parseBool
		}
		break
	case OPTIONS_REWRITE:
		// nginx.ingress.kubernetes.io/rewrite-target: /$1 —— 代表取第一块内容，也就是(.*)
		// host:acc.com  path:/abc/(.*) 这里的第一块是通配符
		// 通过访问acc.com/abc/访问  //如果后端服务是/login，就通过acc.com/abc/login访问过去
		if value, ok := ingress.ObjectMeta.Annotations[ANNOTATIONS_REWRITE]; ok {
			parseBool, _ := strconv.ParseBool(value)
			return parseBool
		}
		break
	}
	return false
}

func (this *IngressSvc) PostIngress(cluster string, post *IngressPost) error {
	//ingress类型
	className := "nginx"
	pathType := v1.PathTypePrefix
	var ingressRules []v1.IngressRule
	// 凑 Rule对象
	for _, r := range post.Rules {
		httpRuleValue := &v1.HTTPIngressRuleValue{}
		rulePaths := make([]v1.HTTPIngressPath, 0)
		for _, pathCfg := range r.Paths {
			port, err := strconv.Atoi(pathCfg.Port)
			if err != nil {
				return err
			}
			rulePaths = append(rulePaths, v1.HTTPIngressPath{
				Path:     pathCfg.Path,
				PathType: &pathType,
				Backend: v1.IngressBackend{
					Service: &v1.IngressServiceBackend{
						Name: pathCfg.SvcName,
						Port: v1.ServiceBackendPort{
							Number: intstr.FromInt(port).IntVal, //这里需要FromInt
						},
					},
				},
			})
		}
		httpRuleValue.Paths = rulePaths
		rule := v1.IngressRule{
			Host: r.Host,
			IngressRuleValue: v1.IngressRuleValue{
				HTTP: httpRuleValue,
			},
		}
		ingressRules = append(ingressRules, rule)
	}

	// 凑 Ingress对象
	ingress := &v1.Ingress{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Ingress",
			APIVersion: "networking.k8s.io/v1beta1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:        post.Name,
			Namespace:   post.Namespace,
			Annotations: this.parseAnnotations(post.Annotations),
		},
		Spec: v1.IngressSpec{
			IngressClassName: &className,
			Rules:            ingressRules,
		},
	}

	_, err := (*this.ClintMap)[cluster].NetworkingV1().Ingresses(post.Namespace).
		Create(context.Background(), ingress, metav1.CreateOptions{})
	return err
}

// 解析annotations标签
func (this *IngressSvc) parseAnnotations(annos string) map[string]string {
	replace := []string{"\t", " ", "\n", "\r\n"}
	for _, r := range replace {
		annos = strings.ReplaceAll(annos, r, "")
	}
	ret := make(map[string]string)
	list := strings.Split(annos, ";")
	for _, item := range list {
		annos := strings.Split(item, ":")
		if len(annos) == 2 {
			ret[annos[0]] = annos[1]
		}
	}
	return ret

}

func (this *IngressSvc) DeleteIngress(cluster, ns, name string) error {
	err := (*this.ClintMap)[cluster].NetworkingV1().Ingresses(ns).Delete(context.Background(), name, metav1.DeleteOptions{})
	return err
}
