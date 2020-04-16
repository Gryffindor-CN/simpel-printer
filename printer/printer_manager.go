package printer

type Manager interface {

	Add(name *string, device *string) error

	Delete(name *string)

	List(added bool) (*List, error)

	Print(printInfo *PrintInfo) (*PrintResult, error)

	Job(printer *string, jobId *string) (*JobInfo, error)

	JobList(printer *string, status *string) (*JobInfoList, error)
}

type List struct {
	Printers []Printer `json:"printers"`
}

type Printer struct {
	Name string `json:"name"`
	StatusCn string `json:"statusCn"`
	StatusEn string `json:"statusEn"`
	Device string `json:"device"`
	Connected bool `json:"connected"`
	Supported bool `json:"supported"`
}

type JobInfoList struct {
	Jobs []JobInfo `json:"jobs"`
}

type JobInfo struct {
	Id string `json:"id"`
	StartTime string `json:"startTime"`
	Status string `json:"status"`
	Description string `json:"description"`
	FileSize int64 `json:"fileSize"`
}

type PrintInfo struct {
	//文件下载地址
	Url string `json:"url"`
	//打印机名称
	Printer string `json:"printer"`
	//纸张宽（cm）
	Width string `json:"width"`
	//纸张高（cm）
	Height string `json:"height"`
	//数量
	Quantity string `json:"quantity"`

}

type PrintResult struct {
	JobId string `json:"jobId"`
}
