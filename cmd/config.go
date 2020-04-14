package cmd

type config struct {
	DefaultPorts     []int
	DefaultQueryPort int
	DefaultUsername  string
	DefaultPassword  string
	Violators        string
	Timelimit        int
	Kicklimit        int
	Action           string
	Message          string
	KickMessage      string
	BanMessage       string
	BanDuration      int
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
