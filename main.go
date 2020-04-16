package main

import (
	"./common"
	"./service"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func ServeHttp(w http.ResponseWriter, r *http.Request) {
	var host = "gitsvr.mipesoft.com"
	remote, err := url.Parse("http://" + host + r.URL.Path)
	if err != nil {
		panic(err)
	}
	remote.RawQuery = r.URL.RawQuery
	proxy := httputil.NewSingleHostReverseProxy(remote)
	r.Host = host
	r.Header.Set("Host", host)
	proxy.ServeHTTP(w, r)
}

func PingHttp(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("pong"))
}

func startServer() {
	http.HandleFunc("/", ServeHttp)
	http.HandleFunc("/ping", PingHttp)
	http.HandleFunc("/printer/add", service.AddPrinter)
	http.HandleFunc("/printer/list", service.ListPrinter)
	http.HandleFunc("/printer/print", service.Print)
	http.HandleFunc("/printer/job", service.Job)
	http.HandleFunc("/printer/job/list", service.JobList)
	http.HandleFunc("/device/logs", service.GetLogs)
	http.ListenAndServe(":8888", nil)
}

func main() {
	// 日志
	//common.Log.WithFields(logrus.Fields{
	//	"animal": "monkey",
	//	"size": 10,
	//}).Error("A group of walrus emerges from the ocean")
	//common.Log.Info("info test.")

	//var bootStrap service.Bootstrap = new (service.LanCable)
	//bootStrap.Start()
	startServer()
}

func init()  {
	common.InitLog()
}