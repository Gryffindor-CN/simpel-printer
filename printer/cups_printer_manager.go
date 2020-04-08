package printer

import (
	"fmt"
	"os/exec"
	"strings"
)

type /**/CupsManager struct {
	
}

func (cupsManager CupsManager) Add() {
	var (
		cmd *exec.Cmd
		output []byte
		err error
	)

	cmd = exec.Command("/bin/bash", "-c", "lp -o media=Custom.4x7cm 4-7.pdf -d gt888k")

	if output, err = cmd.CombinedOutput(); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(string(output))
}

func (cupsManager CupsManager) List() List {

	list := List{printers:nil}

	// 执行查询打印机列表命令
	var (
		output []byte
		err error
	)
	cmd := exec.Command("/bin/bash", "-c", "ssh root@192.168.206.115 lpstat -p")
	if output, err = cmd.CombinedOutput(); err != nil {
		fmt.Println(err)
		return list
	}

	// 解析返回结果
	results := string(output)
	resultArr := strings.Split(results, "\n")
	// 遍历打印机列表
	for i:= 0;i<len(resultArr);i++{
		if resultArr[i] == "" {
			continue
		}

		// 解析打印机信息
		metaArr := strings.Fields(resultArr[i])
		var status string
		if metaArr[2] == "目前空闲。从" {
			status = "空闲"
		}
		list.printers = append(list.printers, Printer{name:metaArr[1],status:status})
	}

	return list
}


func (cupsManager CupsManager) Print() {

}

func (cupsManager CupsManager) Job() {

}

func (cupsManager CupsManager) JobList() {

}

