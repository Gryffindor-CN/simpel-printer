package service

import (
	"../printer"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

func AddPrinter(writer http.ResponseWriter, request *http.Request) {

	// 获取body
	reqBody, _ :=ioutil.ReadAll(request.Body)
	var printerReq PrinterReq
	json.Unmarshal(reqBody, &printerReq)

	// 添加打印机
	var cups printer.Manager = new (printer.CupsManager)
	err := cups.Add(&printerReq.Printer, &printerReq.Device)
	if err != nil {
		handelErr(err, writer)
		return
	}

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
	var added *bool
	for key, value := range form {
		if key == "added" {
			added = new(bool)
			*added, _ = strconv.ParseBool(value[0])
		}
	}
	if added == nil{
		handelErr(errors.New("缺少参数：added"), writer)
		return
	}

	// 获取打印机列表
	var cups printer.Manager = new (printer.CupsManager)
	list, err := cups.List(*added)
	if err != nil {
		handelErr(err, writer)
		return
	}

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

	// 校验参数
	if printInfo.Url == "" || printInfo.Printer == "" || printInfo.Height == "" || printInfo.Width == "" {
		handelErr(errors.New("缺少参数"), writer)
		return
	}

	// 执行打印
	var cups printer.Manager = new (printer.CupsManager)
	res, err := cups.Print(&printInfo)
	if err != nil {
		handelErr(err, writer)
		return
	}

	// response
	str, _ := json.Marshal(res)
	writer.Header().Set("Content-Type","application/json")
	writer.Write(str)
}

func Job(writer http.ResponseWriter, request *http.Request) {

	// 获取参数
	request.ParseForm()
	form := request.Form
	var printerName, jobId *string
	for key, value := range form {
		if key == "printer" {
			printerName = new(string)
			*printerName = value[0]
		}
		if key == "jobId" {
			jobId = new(string)
			*jobId = value[0]
		}
	}

	// 参数校验
	if printerName == nil || jobId == nil {
		handelErr(errors.New("缺少参数"), writer)
		return
	}

	// 执行打印
	var cups printer.Manager = new (printer.CupsManager)
	// 查询单个任务
	job, err := cups.Job(printerName, jobId)
	if err != nil {
		handelErr(err, writer)
		return
	}

	// response
	str, _ := json.Marshal(job)
	writer.Header().Set("Content-Type","application/json")
	writer.Write(str)
}

func JobList(writer http.ResponseWriter, request *http.Request) {

	// 获取参数
	request.ParseForm()
	form := request.Form

	var printerName, status *string
	for key, value := range form {
		if key == "printer" {
			printerName = new(string)
			*printerName = value[0]
		}
		if key == "status" {
			status = new(string)
			*status = value[0]
		}
	}

	// 参数校验
	if printerName == nil || status == nil {
		handelErr(errors.New("缺少参数"), writer)
		return
	}

	// 执行打印
	var cups printer.Manager = new (printer.CupsManager)
	// 查询单个任务
	jobList, err := cups.JobList(printerName, status)
	if err != nil {
		handelErr(err, writer)
		return
	}

	// response
	str, _ := json.Marshal(jobList)
	writer.Header().Set("Content-Type","application/json")
	writer.Write(str)
}

func handelErr(err error, writer http.ResponseWriter)  {
	var error PrinterError
	error.Code = "99"
	error.Message = err.Error()
	str, _ := json.Marshal(error)
	writer.Header().Set("Content-Type","application/json")
	writer.Write(str)
}

type PrinterError struct {
	Code string `json:"code"`
	Message string `json:"message"`
}
