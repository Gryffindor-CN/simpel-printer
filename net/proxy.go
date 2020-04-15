package net

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
)

type Proxy interface {
	Register() string
}

type PortwayProxy struct {
	tunnel string
	agentPath string
}

const(
	tag = "service"
)

func NewPortwayProxy(tunnel string, agentPath string) *PortwayProxy {
	return &PortwayProxy{tunnel: tunnel, agentPath: agentPath}
}

// 注册httpMap,并返回endpoint
func (portway *PortwayProxy) Register() string {
	// 生成配置文件
	doConfigure(portway.tunnel, portway.agentPath)

	// 启动agent
	if err := startup(portway.agentPath); err != nil {
		log.Println("[x]", "Portway agent 启动失败：" + err.Error())
	}

	var bts bytes.Buffer
	bts.WriteString(tag)
	bts.WriteString("___")
	bts.WriteString("iotube-")
	bts.WriteString(portway.tunnel)
	bts.WriteString(".")
	bts.WriteString("iotube.net")
	return bts.String()
}

// 生成配置文件 agent.ini
func doConfigure(tunnel string, configPath string) {
	var(
		cmd        *exec.Cmd
		err        error
		in         *bytes.Buffer
		out        bytes.Buffer
		agent_init = " << EOF\n" +
			"[setting]\n" +
			"center_host=gw.iotube.net\n" +
			"center_webapi=http://info.iotube.net\n" +
			"center_port=80\n" +
			"tunnel=" + "iotube-" + tunnel + "\n" +
			"apiport=27093\n" +
			"apipath=/api/v1\n" +
			"token=\n" +
			"service_name=\n" +
			"service_disp=\n" +
			"service_desc=\n\n" +
			"[httpmap]\n" +
			tag + "=127.0.0.1:8888\n" +
			"java=127.0.0.1:8080\n" +
			"EOF"
	)
	cmd = exec.Command("/bin/bash")
	in = bytes.NewBuffer(nil)
	cmd.Stdin = in
	cmd.Stdout = &out
	go func() {
		//in.WriteString("ssh root@192.168.206.115\n") //TODO 删除调试行
		in.WriteString("cat > " + configPath + "/agent.ini" + agent_init)
	}()
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
		panic(err)
	}
	if err = cmd.Wait(); err != nil {
		fmt.Println("Command finished with error: %v", err)
		panic(err)
	}
}

func startup(agentPath string) error {
	var(
		err error
		out bytes.Buffer
		in = bytes.NewBuffer(nil)
		filePath = agentPath + "/agent"
	)

	//TODO 部署时必须开放代码
	fileIsExist, err := pathExists(filePath)
	if err != nil  {
		return err
	}
	if !fileIsExist {
		return errors.New("找不到文件[" + filePath + "]")
	}

	cmd := exec.Command("/bin/bash")
	cmd.Stdin = in
	cmd.Stdout = &out
	go func() {
		//in.WriteString("ssh root@192.168.206.115\n") //TODO 删除调试行
		in.WriteString("nohup " + filePath + " > /dev/null 2>&1 &")
	}()
	if err = cmd.Start(); err != nil {
		log.Fatal(err)
		panic(err)
	}
	if err = cmd.Wait(); err != nil {
		fmt.Println("Command finished with error: %v", err)
		panic(err)
	}
	log.Println(out.String())
	return nil
}

func pathExists(path string) (isExist bool, error error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}