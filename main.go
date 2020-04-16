package main

import (
	"./common"
	"github.com/sirupsen/logrus"
)

func main() {
	common.Log.WithFields(logrus.Fields{
		"animal": "monkey",
		"size": 10,
	}).Error("A group of walrus emerges from the ocean")
	common.Log.WithFields(logrus.Fields{
		"name": "cp",
	}).Warn("hello world.")
	common.Log.Info("info test.")
}

func init()  {
	common.InitLog()
}





