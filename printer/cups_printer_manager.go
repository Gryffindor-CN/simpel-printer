package printer

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

type CupsManager struct {
	
}

func (cupsManager CupsManager) Add(name *string, device *string) {
	usbPrinter := getUsbPrinter()

	if device == nil || name == nil || usbPrinter == nil {
		return
	}

	if *device != *usbPrinter {
		return
	}

	driver := getPrinterDriver(device)

	if driver == nil {
		return
	}

	command := "ssh root@192.168.206.115 'lpadmin -p " + *name + " -v \"" + *device + "\" -m \"" + *driver + "\"'"
	exeCommand(command)
	exeCommand("ssh root@192.168.206.115 cupsenable " + *name)
	exeCommand("ssh root@192.168.206.115 cupsaccept " + *name)
}

func (cupsManager CupsManager) List(added *bool) *List {

	if *added {
		return addedList()
	} else {
		return notAddedList()
	}
}

func (cupsManager CupsManager) Print() {

}

func (cupsManager CupsManager) Job(printer *string, jobId *string) *JobInfo {

	var jobInfo JobInfo

	if jobId == nil || *jobId == "" {
		return nil
	}

	status := "all"

	jobList := cupsManager.JobList(printer, &status)

	if jobList == nil {
		return nil
	}
	for i:= 0; i<len(jobList.jobs); i++ {
		job := jobList.jobs[i]
		if job.id == *jobId {
			jobInfo = job
		}
	}

	return &jobInfo
}

func (cupsManager CupsManager) JobList(printer *string, status *string) *JobInfoList {

	if printer == nil || *printer == "" || status == nil || *status == "" {
		return nil
	}

	jobList := JobInfoList{jobs:nil}

	results := exeCommand("ssh root@192.168.206.115 lpstat -W " + *status + " -l -o " + *printer)
	results = strings.Replace(results, "\t", "", -1)

	resultArr := strings.Split(results, "列队\n")

	for i:= 0; i<len(resultArr); i++ {
		if resultArr[i] == "" {
			continue
		}
		var jobInfo JobInfo

		jobArr := strings.Split(resultArr[i], "\n")
		//获取任务ID、名称、文件大小、开始时间
		metaArr := strings.Fields(jobArr[0])
		jobInfo.id = metaArr[0]
		if fileSize, err := strconv.ParseInt(metaArr[2], 10, 32); err != nil {
			fmt.Println(err)
			jobInfo.fileSize = 0
		} else {
			jobInfo.fileSize = fileSize
		}
		time := metaArr[3]
		time += metaArr[5]
		jobInfo.startTime = time

		//获取状态
		statusStr := jobArr[2]
		statusStr = strings.Replace(statusStr, "警告：", "", -1)
		jobInfo.status = str2JobStatus(statusStr)

		//获取描述
		description := jobArr[1]
		description = strings.Replace(description, "状态：", "", -1)
		jobInfo.description = description

		jobList.jobs = append(jobList.jobs, jobInfo)
	}

	return &jobList
}

/**
 * 获取已连接的打印机列表
 */
func addedList() *List  {
	results := exeCommand("ssh root@192.168.206.115 lpstat -p")

	if results == "" {
		return nil
	}

	usbPrinter := getUsbPrinter()

	list := List{printers:nil}
	resultArr := strings.Split(results, "\n")
	// 遍历打印机列表
	for i:= 0;i<len(resultArr);i++{
		if resultArr[i] == "" || resultArr[i] == "\t未知原因" || resultArr[i] == "\tPaused" || resultArr[i] == "\tWaiting for printer to become available." {
			continue
		}
		var printer Printer

		// 解析打印机信息
		metaArr := strings.Fields(resultArr[i])

		printer.supported = true
		printer.status = "unknown"

		if metaArr[2] == "目前空闲。从" {
			printer.status = "空闲"
		}
		if metaArr[2] == "正在打印" {
			printer.status = "打印中"
		}
		if metaArr[6] == "开始被禁用" {
			printer.status = "禁用"
		}
		printer.name = metaArr[1]

		// 判断是否已连接
		connectInfo := getConnectInfoByName(&printer.name)

		if usbPrinter != nil {
			printer.connected = *connectInfo == *usbPrinter
		}
		printer.device = *connectInfo

		list.printers = append(list.printers, printer)
	}

	return &list
}

/**
 * 获取未连接的打印机列表
 */
func notAddedList() *List  {
	results := exeCommand("ssh root@192.168.206.115 lpinfo -v")

	if results == "" {
		return nil
	}

	list := List{printers:nil}
	resultArr := strings.Split(results, "\n")
	connectedList := getConnectInfoList()
	for _, result := range resultArr {
		if result == "" {
			continue
		}

		if !strings.HasPrefix(result, "direct usb://") {
			continue
		}

		device := strings.Fields(result)
		connected := false

		//判断是已连接的，跳过
		if connectedList != nil {
			for _, value := range *connectedList {
				if value == device[1] {
					connected = true
				}
			}
		}

		if connected {
			continue
		}

		var printer Printer
		printer.name = "unknow"
		printer.status = "unknow"
		printer.device = device[1]
		printer.connected = true
		printer.supported = checkPrinterSupport(&printer.device)

		list.printers = append(list.printers, printer)
	}

	return &list
}

/**
 * 获取用usb端口连接的打印机
 */
func getUsbPrinter() *string {
	results := exeCommand("ssh root@192.168.206.115 lpinfo -v")

	if results == "" {
		return nil
	}

	resultArr := strings.Split(results, "\n")

	for i:= 0;i<len(resultArr);i++{
		if resultArr[i] == "" {
			continue
		}

		if strings.HasPrefix(resultArr[i], "direct usb://") {
			device := strings.Fields(resultArr[i])
			return &device[1]
		}
	}

	return nil
}

/**
 * 根据打印机名称获取连接信息
 */
func getConnectInfoByName(name *string) *string  {

	if name == nil {
		return nil
	}

	results := exeCommand("ssh root@192.168.206.115 lpstat -v")

	if results == "" {
		return nil
	}

	resultArr := strings.Split(results, "\n")

	for i:= 0;i<len(resultArr);i++{
		if resultArr[i] == "" {
			continue
		}

		device := strings.Fields(resultArr[i])

		if device[1] == *name {
			connectInfo := strings.Replace(device[2], "的设备：", "", -1)
			return &connectInfo
		}

	}
	return nil
}

/**
 * 获取已添加的打印机连接信息列表
 */
func getConnectInfoList() *[]string  {

	results := exeCommand("ssh root@192.168.206.115 lpstat -v")

	if results == "" {
		return nil
	}

	resultArr := strings.Split(results, "\n")

	var list []string

	for i:= 0;i<len(resultArr);i++{
		if resultArr[i] == "" {
			continue
		}

		device := strings.Fields(resultArr[i])

		list = append(list, strings.Replace(device[2], "的设备：", "", -1))

	}
	return &list
}

/**
 * 执行命令并返回输出
 */
func exeCommand(command string) string {
	// 执行命令
	var (
		output []byte
		err error
	)
	cmd := exec.Command("/bin/bash", "-c", command)
	if output, err = cmd.CombinedOutput(); err != nil {
		fmt.Println(err)
		return ""
	}

	// 解析返回结果
	return string(output)
}

/**
 * 转换任务状态字符串
 */
func str2JobStatus(str string) string {
	switch str {
	case "job-canceled-by-user": return "canceled"
	case "processing-to-stop-point": return "done"
	case "job-printing": return "printing"
	default: return "unknown"
	}
}

/**
 * 校验打印机是否被支持
 */
func checkPrinterSupport(device *string) bool  {
	driver := getPrinterDriver(device)

	if driver == nil {
		return false
	} else {
		return true
	}
}

/**
 * 根据连接信息获取驱动信息
 */
func getPrinterDriver(device *string) *string  {
	deviceArr := strings.Split(*device, "?")

	var driver string

	switch deviceArr[0] {
		case "usb://Zebra%20Technologies/ZTC%20GK888t%20(EPL)":
			driver = "drv:///sample.drv/zebraep2.ppd"
			break
		default:
			return nil
	}

	return &driver
}

