package main

import (
	"./common"
	"./service"
	//"github.com/sirupsen/logrus"
)

func main() {
	// 日志
	//common.Log.WithFields(logrus.Fields{
	//	"animal": "monkey",
	//	"size": 10,
	//}).Error("A group of walrus emerges from the ocean")
	//common.Log.Info("info test.")

	var bootStrap service.Bootstrap = new (service.LanCable)
	bootStrap.Start()
}

func init()  {
	common.InitLog()
}