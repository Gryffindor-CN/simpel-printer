package main

import (
	"./service"
)

func main() {
	var bootStrap service.Bootstrap = new (service.LanCable)
	bootStrap.Start()
}