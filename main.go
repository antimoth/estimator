package main

import (
	"selfconf"

	"github.com/antimoth/estimator/estimator"
)

func RunEstimator() {
	conn, err := estimator.NewEstimator()
	if err != nil {
		return
	}
	defer conn.Close()

	// 启动监听
	conn.StartWatch(selfconf.EstimatorStart)
	conn.RunEstimator()

	ch := make(chan struct{})
	<-ch
}

func main() {
	RunEstimator()
}
