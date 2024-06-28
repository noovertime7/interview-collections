package utils

import (
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
	"log"
)

func HandleCommand(ns, pod, container string, client *kubernetes.Clientset, config *rest.Config, command []string) remotecommand.Executor {
	option := &v1.PodExecOptions{
		Container: container,
		Command:   command,
		//如果是一次性执行命令"sh -c ls"这种就关闭stdin（打开也不影响）
		//如果要从reader里获取"sh"，需要前端的远程shell终端交互的，一定要开启
		Stdin:  true,
		Stdout: true,
		Stderr: true,
		//终端
		TTY: true,
	}
	//构建一个远程命令执行的请求
	req := client.CoreV1().RESTClient().Post().Resource("pods").
		Namespace(ns).
		Name(pod).
		SubResource("exec").
		Param("color", "false").
		VersionedParams(
			option,
			scheme.ParameterCodec,
		)

	//创建远程命令执行对象
	exec, err := remotecommand.NewSPDYExecutor(config, "POST",
		req.URL())
	if err != nil {
		log.Fatal(err)
	}
	return exec
}
