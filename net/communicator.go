package net

// 定义通讯接口
type communicator interface {
	// 接入服务端
	Start() string
	// 关闭接入
	Stop() string
	// 机器注册（绑定账户）
	Registr() string

}