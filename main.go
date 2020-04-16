package main

import (
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
	http.HandleFunc("/printer/delete", service.DeletePrinter)
	http.ListenAndServe(":8888", nil)
}

func main() {
	startServer()
}
