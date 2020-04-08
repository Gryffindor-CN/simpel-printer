package printer

type Manager interface {

	Add()

	List() List

	Print()

	Job()

	JobList()
}

type List struct {
	printers []Printer
}

type Printer struct {
	name string
	status string
}

