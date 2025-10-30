package casbinauth

type Policy struct {
	Subject   string
	Domain    string
	Object    string
	Action    string
	Condition string
}

type GroupingPolicy struct {
	Subject      string
	GroupSubject string
	Domain       string
}
