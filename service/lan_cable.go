package service

import (
	"../common"
	_net "../net"
	"bufio"
	"encoding/json"
	"github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"strconv"
	"strings"
)

const(
	AGENT_PATH string = "./portway_arm"
	DEVICE_SERVER_ENDPOINT string = "http://iot-device.dev.iotube.net/devices"
)

//网线版设备接入
type LanCable struct {
}

func (lanCable *LanCable) Start() string {
	var(
		serial string
		proxy  _net.Proxy
	)

	// 获得机器码
	serial = getSerial()

	// 打开agent，注册httpmap
	proxy = _net.NewPortwayProxy(serial, AGENT_PATH)
	endpoint := proxy.Register()

	//TODO 向设备服务发送接入信息
	registerDevice(serial, "http://" + endpoint)
	return ""
}

// 获得机器码
func getSerial() string {
	var(
		cmd    *exec.Cmd
		output []byte
		err    error
		str    string
	)
	cmd = exec.Command("/bin/bash", "-c", "cat /sys/class/net/eth0/address")
	if output, err = cmd.CombinedOutput(); err != nil {
		panic(err)
	}
	str = strings.Replace(string(output), "\n", "", -1)
	str = strings.Replace(str, ":", "", -1)
	return str
}

type Response struct {
	Code string `json:"code"`
	Message string `json:"message"`
	RequestId string `json:"requestId"`
	Result interface{} `json:"result"`
}

func registerDevice(serial string, endpoint string) {
	data := map[string]string{"serial": serial, "endpoint": endpoint}
	mjson, _ := json.Marshal(data)
	body := strings.NewReader(string(mjson))
	resp, err := http.Post(DEVICE_SERVER_ENDPOINT, "application/json", body)
	if err != nil {
		//TODO http post 请求异常处理
	}
	defer resp.Body.Close();
	bodyStr,_ := ioutil.ReadAll(resp.Body)
	var response Response
	if err = json.Unmarshal(bodyStr, &response); err != nil {
		//TODO json 格式化异常处理
	}

	switch response.Code {
		case "0":
			common.Log.WithFields(logrus.Fields{"serial": serial,"endpoint": endpoint}).Info("[√] 成功接入设备服务")
		default:
			common.Log.WithFields(logrus.Fields{"message": response.Message}).Error("[x] 设备服务接入失败")
	}
}

func printLog(page, size int) (logs []string, err error) {
	var(
		cmd *exec.Cmd
		outputs []string
	)
	start := (page - 1) * size + 1
	end := page * size
	script := "awk 'NR>=" + strconv.Itoa(start) + " && NR <=" + strconv.Itoa(end) + "' ./simple-printer.log"
	cmd = exec.Command("/bin/bash", "-c", script)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	cmd.Start()
	reader := bufio.NewReader(stdout)
	for {
		line, err2 := reader.ReadString('\n')
		if err2 != nil || io.EOF == err2 {
			break
		}
		line = strings.Replace(line, "\n", "", -1)
		outputs = append(outputs, line)
	}
	cmd.Wait()
	return outputs, nil
}