package printer

type Manager interface {

	Add(name *string, device *string)

	List(added *bool) *List

	Print(printInfo *PrintInfo) *PrintResult

	Job(printer *string, jobId *string) *JobInfo

	JobList(printer *string, status *string) *JobInfoList
}

type List struct {
	Printers []Printer
}

type Printer struct {
	Name string
	Status string
	Device string
	Connected bool
	Supported bool
}

type JobInfoList struct {
	Jobs []JobInfo
}

type JobInfo struct {
	Id string
	StartTime string
	Status string
	Description string
	FileSize int64
}

type PrintInfo struct {
	//文件下载地址
	Url string
	//打印机名称
	Printer string
	//纸张宽（cm）
	Width string
	//纸张高（cm）
	Height string

}



type PrintResult struct {
	JobId string
}
