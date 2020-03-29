package cmd

type config struct {
	DefaultPorts     []int
	DefaultQueryPort int
	DefaultUsername  string
	DefaultPassword  string
	Timelimit        int
	Action           string
	Message          string
	Delay            int
	IgnoreGroupNames []string
	Servers          []server
}

type server struct {
	IP        string
	QueryPort int
	Ports     []int
	Username  string
	Password  string
}
