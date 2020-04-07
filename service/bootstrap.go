package service

// 定义入口程序接口
type Bootstrap interface {
	// 接入服务端
	Start() string
}