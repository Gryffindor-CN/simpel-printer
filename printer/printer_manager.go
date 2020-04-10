package printer

type Manager interface {

	Add(name *string, device *string)

	List(added *bool) *List

	Print()

	Job(printer *string, jobId *string) *JobInfo

	JobList(printer *string, status *string) *JobInfoList
}

type List struct {
	printers []Printer
}

type Printer struct {
	name string
	status string
	device string
	connected bool
	supported bool
}

type JobInfoList struct {
	jobs []JobInfo
}

type JobInfo struct {
	id string
	startTime string
	status string
	description string
	fileSize int64
}

