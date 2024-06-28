package utils

import (
	"context"
	"fmt"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"k8sapi-pro/src/models"
	"regexp"
)

const DOMAIN_FORMAT = "[a-zA-Z0-9][-a-zA-Z0-9]{0,62}(\\.[a-zA-Z0-9][-a-zA-Z0-9]{0,62})+\\.?"

func IsDomainName(key string) bool {
	return regexp.MustCompile(DOMAIN_FORMAT).MatchString(key)
}

func LabelsFilter(labels map[string]string) (ret []string) {
	for k, v := range labels {
		if !IsDomainName(k) {
			ret = append(ret, fmt.Sprintf("%s=%s", k, v))
		}
	}
	return
}

func TaintsFilter(taints []v1.Taint) (ret []string) {
	for _, taint := range taints {
		if !IsDomainName(taint.Key) {
			ret = append(ret, fmt.Sprintf("%s=%s:%s", taint.Key, taint.Value, taint.Effect))
		}
	}
	return
}

// 第一个是cpu使用 第二个是内存使用
func GetNodeUsage(c *versioned.Clientset, node *v1.Node) []float64 {
	nodeMetric, _ := c.MetricsV1beta1().
		NodeMetricses().Get(context.Background(), node.Name, metav1.GetOptions{})
	cpu := float64(nodeMetric.Usage.Cpu().MilliValue()) / float64(node.Status.Capacity.Cpu().MilliValue())
	memory := float64(nodeMetric.Usage.Memory().MilliValue()) / float64(node.Status.Capacity.Memory().MilliValue())
	return []float64{cpu, memory}
}

// 获取节点配置
func GetNodeConfig(c *models.SysConfig, nodeName string) *models.NodesConfig {
	for _, node := range c.K8s.Nodes {
		if node.Name == nodeName {
			return node
		}
	}
	panic(any("no such node"))
}
