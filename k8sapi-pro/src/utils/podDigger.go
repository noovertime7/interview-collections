package utils

import (
	"fmt"
	v1 "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func GetImages(dep v1.Deployment) string {
	return GetImagesByPod(dep.Spec.Template.Spec.Containers)
}

func GetImagesByPod(containers []core.Container) string {
	images := containers[0].Image
	if imgLen := len(containers); imgLen > 1 {
		images += fmt.Sprintf("+其他%d个镜像", imgLen-1)
	}
	return images
}

func GetPodIsReady(pod *core.Pod) bool {
	//判断pod状态正常
	//1.pod的phase（阶段）值：取Running
	//2.pod的podConditions, 状态值全为true
	if pod.Status.Phase != "Running" {
		return false
	}
	for _, condition := range pod.Status.Conditions {
		if condition.Status != "True" {
			return false
		}
	}
	for _, rg := range pod.Spec.ReadinessGates {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == rg.ConditionType && condition.Status != "True" {
				return false
			}
		}
	}
	return true
}

func IsCompleted(deployment *v1.Deployment) bool {
	return *(deployment.Spec.Replicas) == deployment.Status.AvailableReplicas
}

// GetAvailableMessage 对deployment的状态显示更详细的可以选择从实体.status.conditons数组中获取
// 其中有Available（最小副本可用），Progressing ，ReplicaFailure等种类
// 其对应的Message有详细的信息
func GetAvailableMessage(deployment *v1.Deployment) string {
	for _, v := range deployment.Status.Conditions {
		if string(v.Type) == "Available" {
			return v.Message
		}
	}
	return ""
}

// 快捷创建时  需要 初始化一些 标签
func InitLabel(deploy *v1.Deployment) {
	if deploy.Spec.Selector == nil {
		deploy.Spec.Selector = &v12.LabelSelector{MatchLabels: map[string]string{"app": deploy.Name}}
	}
	if deploy.Spec.Selector.MatchLabels == nil {
		deploy.Spec.Selector.MatchLabels = map[string]string{"app": deploy.Name}
	}
	if deploy.Spec.Template.ObjectMeta.Labels == nil {
		deploy.Spec.Template.ObjectMeta.Labels = map[string]string{"app": deploy.Name}
	}
	deploy.Spec.Selector.MatchLabels["app"] = deploy.Name

	deploy.Spec.Template.ObjectMeta.Labels["app"] = deploy.Name
}
