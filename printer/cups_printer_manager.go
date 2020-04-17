package printer

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type CupsManager struct {
	
}

func (cupsManager CupsManager) Add(name *string, device *string) error {
	// TODO 校验打印机名称是否已存在

	usbPrinter, err := getUsbPrinter()
	if err != nil {
		return nil
	}

	if device == nil || name == nil {
		return errors.New("缺少参数")
	}

	if *device == "" || *name == "" {
		return errors.New("缺少参数")
	}

	if usbPrinter == nil || *usbPrinter == "" || *device != *usbPrinter {
		return errors.New("要添加的打印机未连接")
	}

	driver := getPrinterDriver(device)

	if driver == nil {
		return errors.New("要添加的打印机未被支持")
	}

	command := "lpadmin -p " + *name + " -v \"" + *device + "\" -m \"" + *driver + "\""
	if _, err := exeCommand(command); err != nil {
		return err
	}
	if _, err := exeCommand("cupsenable " + *name); err != nil {
		return err
	}
	if _, err := exeCommand("cupsaccept " + *name); err != nil {
		return err
	}
	return nil
}

func (cupsManager CupsManager) List(added bool) (*List, error) {

	if added {
		return addedList()
	} else {
		return notAddedList()
	}
}

func (cupsManager CupsManager) Print(printInfo *PrintInfo) (*PrintResult, error) {

	// 生成文件名称（时间戳）
	timestamp := time.Now().UnixNano()
	fileName := strconv.FormatInt(timestamp,10) + ".pdf"
	fmt.Println(fileName)

	path := "/data/" + fileName
	//path := "/data/15867490549580640002.pdf"

	// 下载文件
	httpResp, err := http.Get(printInfo.Url)
	if err != nil {
		return nil, errors.New("下载文件失败")
	}
	if httpResp.StatusCode != 200 {
		return nil, errors.New("下载文件失败")
	}
	file, err := os.Create(path)
	if err != nil {
		return nil, errors.New("保存文件失败")
	}
	io.Copy(file, httpResp.Body)
	defer file.Close()

	// 打印文件
	cmd := "lp -o orientation-requested=6 -o print-quality=3 -o media=Custom." + printInfo.Width + "x" + printInfo.Height + "cm -n " + printInfo.Quantity + " -d " + printInfo.Printer + " " + path;
	exeResp, err := exeCommand(cmd)
	if err != nil {
		return nil, err
	}
	//返回job ID
	exeRespArr := strings.Fields(*exeResp)

	var result PrintResult
	result.JobId = strings.Replace(exeRespArr[3], "（1", "", -1)
	return &result, nil

}

func (cupsManager CupsManager) Job(printer *string, jobId *string) (*JobInfo, error) {

	var jobInfo JobInfo

	if jobId == nil || *jobId == "" {
		return nil, errors.New("缺少参数")
	}

	status := "all"

	jobList, err := cupsManager.JobList(printer, &status)
	if err != nil {
		return nil, err
	}

	if jobList == nil {
		return nil, errors.New("找不到当前id为" + *jobId + "任务")
	}
	for i:= 0; i<len(jobList.Jobs); i++ {
		job := jobList.Jobs[i]
		if job.Id == *jobId {
			jobInfo = job
		}
	}

	if jobInfo.Id == "" {
		return nil, errors.New("找不到当前id为" + *jobId + "任务")
	}

	return &jobInfo, nil
}

func (cupsManager CupsManager) JobList(printer *string, status *string) (*JobInfoList, error) {

	if printer == nil || *printer == "" || status == nil || *status == "" {
		return nil, errors.New("缺少参数")
	}

	jobList := JobInfoList{Jobs:nil}

	results, err := exeCommand("lpstat -W " + *status + " -l -o " + *printer)
	if err != nil {
		return nil, err
	}
	*results = strings.Replace(*results, "\t", "", -1)

	resultArr := strings.Split(*results, "列队\n")

	for i:= 0; i<len(resultArr); i++ {
		if resultArr[i] == "" {
			continue
		}
		var jobInfo JobInfo

		jobArr := strings.Split(resultArr[i], "\n")
		//获取任务ID、名称、文件大小、开始时间
		metaArr := strings.Fields(jobArr[0])
		jobInfo.Id = metaArr[0]
		if fileSize, err := strconv.ParseInt(metaArr[2], 10, 32); err != nil {
			fmt.Println(err)
			jobInfo.FileSize = 0
		} else {
			jobInfo.FileSize = fileSize
		}
		time := metaArr[3]
		time += metaArr[5]
		jobInfo.StartTime = time

		//获取状态
		statusStr := jobArr[2]
		statusStr = strings.Replace(statusStr, "警告：", "", -1)
		jobInfo.Status = str2JobStatus(statusStr)

		//获取描述
		description := jobArr[1]
		description = strings.Replace(description, "状态：", "", -1)
		jobInfo.Description = description

		jobList.Jobs = append(jobList.Jobs, jobInfo)
	}

	return &jobList, nil
}

func (cupsManager CupsManager) Delete(name *string)  {
	command := "lpadmin -x " + *name
	exeCommand(command)
}

/**
 * 获取已连接的打印机列表
 */
func addedList() (*List, error)  {
	results, _ := exeCommand("lpstat -p")

	list := List{Printers:[]Printer{}}
	if results == nil || *results == "" {
		return &list, nil
	}

	usbPrinter, err := getUsbPrinter()
	if err != nil {
		return nil, err
	}

	resultArr := strings.Split(*results, "\n")
	// 遍历打印机列表
	for i:= 0;i<len(resultArr);i++{
		if resultArr[i] == "" || resultArr[i] == "\t未知原因" || resultArr[i] == "\tPaused" || resultArr[i] == "\tWaiting for printer to become available." {
			continue
		}
		var printer Printer

		// 解析打印机信息
		metaArr := strings.Fields(resultArr[i])

		printer.Supported = true
		printer.StatusCn = "unknown"
		printer.StatusEn = "未知"

		if metaArr[2] == "目前空闲。从" {
			printer.StatusCn = "空闲"
			printer.StatusEn = "idle"
		}
		if metaArr[2] == "正在打印" {
			printer.StatusCn = "打印中"
			printer.StatusEn = "printing"
		}
		if metaArr[6] == "开始被禁用" {
			printer.StatusCn = "禁用"
			printer.StatusEn = "disable"
		}
		printer.Name = metaArr[1]

		// 判断是否已连接
		connectInfo, err := getConnectInfoByName(&printer.Name)
		if err != nil {
			return nil, err
		}

		if usbPrinter != nil {
			printer.Connected = *connectInfo == *usbPrinter
		}
		printer.Device = *connectInfo

		list.Printers = append(list.Printers, printer)
	}

	return &list, nil
}

/**
 * 获取未连接的打印机列表
 */
func notAddedList() (*List, error)  {
	results, _ := exeCommand("lpinfo -v")

	list := List{Printers:[]Printer{}}

	if results == nil || *results == "" {
		return &list, nil
	}


	resultArr := strings.Split(*results, "\n")
	connectedList, _ := getConnectInfoList()
	
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
		printer.Name = "unknow"
		printer.StatusCn = "unknow"
		printer.StatusEn = "未知"
		printer.Device = device[1]
		printer.Connected = true
		printer.Supported = checkPrinterSupport(&printer.Device)

		list.Printers = append(list.Printers, printer)
	}

	return &list, nil
}

/**
 * 获取用usb端口连接的打印机
 */
func getUsbPrinter() (*string, error) {
	results, err := exeCommand("lpinfo --timeout 1 -v")
	if err != nil {
		return nil, err
	}
	if results == nil || *results == "" {
		return nil,nil
	}

	resultArr := strings.Split(*results, "\n")

	for i:= 0;i<len(resultArr);i++{
		if resultArr[i] == "" {
			continue
		}

		if strings.HasPrefix(resultArr[i], "direct usb://") {
			device := strings.Fields(resultArr[i])
			return &device[1], nil
		}
	}

	return nil, nil
}

/**
 * 根据打印机名称获取连接信息
 */
func getConnectInfoByName(name *string) (*string, error)  {

	if name == nil {
		return nil, errors.New("打印机名称为空")
	}

	results, err := exeCommand("lpstat -v")
	if err != nil {
		return nil, err
	}
	if results == nil || *results == "" {
		return nil, nil
	}

	resultArr := strings.Split(*results, "\n")

	for i:= 0;i<len(resultArr);i++{
		if resultArr[i] == "" {
			continue
		}

		device := strings.Fields(resultArr[i])

		if device[1] == *name {
			connectInfo := strings.Replace(device[2], "的设备：", "", -1)
			return &connectInfo, nil
		}

	}
	return nil, nil
}

/**
 * 获取已添加的打印机连接信息列表
 */
func getConnectInfoList() (*[]string, error)  {

	results, err := exeCommand("lpstat -v")
	if err != nil {
		return nil, err
	}

	if results == nil || *results == "" {
		return nil, nil
	}

	resultArr := strings.Split(*results, "\n")

	var list []string

	for i:= 0;i<len(resultArr);i++{
		if resultArr[i] == "" {
			continue
		}

		device := strings.Fields(resultArr[i])

		list = append(list, strings.Replace(device[2], "的设备：", "", -1))

	}
	return &list, nil
}

/**
 * 执行命令并返回输出
 */
func exeCommand(command string) (*string, error) {
	// 执行命令
	var (
		output []byte
		err error
	)
	//cmd := exec.Command("/bin/bash", "-c", "ssh root@192.168.206.115 '" + command + "'")
	cmd := exec.Command("/bin/bash", "-c", command)
	if output, err = cmd.CombinedOutput(); err != nil {
		if output != nil {
			return nil, errors.New("执行命令失败：" + string(output))
		}
		return nil, errors.New("执行命令失败：")
	}

	// 解析返回结果
	res := string(output)
	return &res, nil
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
			driver = "drv:///sample.drv/zebra.ppd"
			break
		default:
			// TODO 输出日志
			return nil
	}

	return &driver
}

