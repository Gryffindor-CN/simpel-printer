package service

import (
	"../printer"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

/**
 * @api {POST} /printer/add 添加打印机
 * @apiGroup printer
 * @apiName printerAdd
 * @apiDescription 添加打印机
 *
 * @apiParam {String} printer 打印机名称
 * @apiParam {String} device 设备信息
 *
 * @apiParamExample 请求示例
 * POST:/printer/add
 * {
 *	"printer":"gk888t6",
 *	"device":"usb://Zebra%20Technologies/ZTC%20GK888t%20(EPL)?serial=19J193906076"
 * }
 *
 * @apiSuccess {String} code 返回码
 * @apiSuccess {String} message 返回消息
 * @apiSuccess {Object} result 返回结果
 *
 * @apiSuccessExample 正确时的返回JSON数据包如下
 * {
 *     "code": "0",
 *     "message": "OK",
 *     "result": null
 * }
 *
 * @apiError printer.99 测试错误
 * @apiErrorExample 错误时的返回JSON数据包如下（示例为缺少参数）
 * {
 *     "code": "printer.99",
 *     "message": "测试错误"
 * }
 */
func AddPrinter(writer http.ResponseWriter, request *http.Request) {

	if request.Method != "POST" {
		handelErr(errors.New("不支持的方法：" + request.Method), writer)
		return
	}

	// 获取body
	reqBody, _ := ioutil.ReadAll(request.Body)
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
	handelResp(nil, writer)
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
	handelResp(list, writer)
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

	handelResp(res, writer)
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

	handelResp(job, writer)
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

	// 查询单个任务
	var cups printer.Manager = new (printer.CupsManager)
	jobList, err := cups.JobList(printerName, status)
	if err != nil {
		handelErr(err, writer)
		return
	}

	handelResp(jobList, writer)
}

func handelErr(err error, writer http.ResponseWriter)  {
	var error PrinterError
	error.Code = "printer.99"
	error.Message = err.Error()
	str, _ := json.Marshal(error)
	writer.Header().Set("Content-Type","application/json")
	writer.Write(str)
}

func handelResp(obj interface{}, writer http.ResponseWriter)  {
	var response PrinterResponse
	// response
	response.Code = "0"
	response.Message = "OK"
	response.Result = obj
	str, _ := json.Marshal(response)
	writer.Header().Set("Content-Type","application/json")
	writer.Write(str)
}

type PrinterError struct {
	Code string `json:"code"`
	Message string `json:"message"`
}

type PrinterResponse struct {
	Code string `json:"code"`
	Message string `json:"message"`
	Result  interface{}`json:"result"`
}