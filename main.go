package main

import (
	"./service/printer"
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

func startServer() {
	http.HandleFunc("/", ServeHttp)
	http.HandleFunc("/printer/add", service_printer.AddPrinter)
	http.HandleFunc("/printer/list", service_printer.ListPrinter)
	http.HandleFunc("/printer/print", service_printer.Print)
	http.HandleFunc("/printer/job", service_printer.Job)
	http.HandleFunc("/printer/job/list", service_printer.JobList)
	http.ListenAndServe(":8888", nil)
}

func main() {
	startServer()
}
