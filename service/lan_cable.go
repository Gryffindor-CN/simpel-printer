package service

import (
	_net "../net"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
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
	fmt.Println(response)

	switch response.Code {
	case "0":
		log.Printf("成功接入设备服务：%s，您可以使用：%s 访问", serial, endpoint)
	default:
		log.Println("设备服务接入失败：" + response.Message)
	}
}