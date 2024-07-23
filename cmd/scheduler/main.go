package main

import (
	"os"

	"k8s.io/kubernetes/cmd/kube-scheduler/app"

	"github.com/helen-frank/mix-scheduler-plugins/pkg/spot"
)

func main() {
	command := app.NewSchedulerCommand(
		app.WithPlugin(spot.Name, spot.New),
	)

	if err := command.Execute(); err != nil {
		os.Exit(1)
	}
}
