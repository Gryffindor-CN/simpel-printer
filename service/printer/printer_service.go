package service_printer

import (
	"../../printer"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
)

func AddPrinter(writer http.ResponseWriter, request *http.Request) {

	// 获取body
	reqBody, err :=ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	var printerReq PrinterReq
	json.Unmarshal(reqBody, &printerReq)

	// 添加打印机
	var cups printer.Manager = new (printer.CupsManager)
	cups.Add(&printerReq.Printer, &printerReq.Device)

	// response
	str := "{\"code\":\"0\",\"message\":\"ok\"}"
	var body = []byte(str)
	writer.Header().Set("Content-Type","application/json")
	writer.Write(body)
}

type PrinterReq struct {
	Printer string `json:"printer"`
	Device string `json:"device"`
}

func ListPrinter(writer http.ResponseWriter, request *http.Request) {

	// 获取参数
	request.ParseForm()
	form := request.Form
	var added bool
	for key, value := range form {
		if key == "added" {
			added, _ = strconv.ParseBool(value[0])
		}
	}

	// 获取打印机列表
	var cups printer.Manager = new (printer.CupsManager)
	list := cups.List(added)

	// response
	str, _ := json.Marshal(list)
	writer.Header().Set("Content-Type","application/json")
	writer.Write(str)
}

func Print(writer http.ResponseWriter, request *http.Request) {

	// 获取body
	reqBody, err :=ioutil.ReadAll(request.Body)
	if err != nil {
		panic(err)
	}
	var printInfo printer.PrintInfo
	json.Unmarshal(reqBody, &printInfo)

	// 执行打印
	var cups printer.Manager = new (printer.CupsManager)
	res := cups.Print(&printInfo)

	// response
	str, _ := json.Marshal(res)
	writer.Header().Set("Content-Type","application/json")
	writer.Write(str)
}

func Job(writer http.ResponseWriter, request *http.Request) {

	// 获取参数
	request.ParseForm()
	form := request.Form
	printerName := form["printer"][0]
	jobId := form["jobId"][0]

	// 执行打印
	var cups printer.Manager = new (printer.CupsManager)
	// 查询单个任务
	//printerName := "gt888k"
	//jobId := "gt888k-68"
	job := cups.Job(&printerName, &jobId)

	// response
	str, _ := json.Marshal(job)
	writer.Header().Set("Content-Type","application/json")
	writer.Write(str)
}

func JobList(writer http.ResponseWriter, request *http.Request) {

	// 获取参数
	request.ParseForm()
	form := request.Form
	printerName := form["printer"][0]
	status := form["status"][0]

	// 执行打印
	var cups printer.Manager = new (printer.CupsManager)
	// 查询单个任务
	jobList := cups.JobList(&printerName, &status)

	// response
	str, _ := json.Marshal(jobList)
	writer.Header().Set("Content-Type","application/json")
	writer.Write(str)
}