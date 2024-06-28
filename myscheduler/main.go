package main

import (
	"fmt"
	"k8s.io/component-base/logs"
	"k8s.io/kubernetes/cmd/kube-scheduler/app"
	"myscheduler/lib"
	"os"
)

// 参考项目 https://github.com/kubernetes-sigs/scheduler-plugins
func main() {
	//来自/blob/master/cmd/scheduler/main.go
	command := app.NewSchedulerCommand(
		app.WithPlugin(lib.TestSchedulingName, lib.NewTestScheduling),
	)
	logs.InitLogs()
	defer logs.FlushLogs()

	if err := command.Execute(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}

}
