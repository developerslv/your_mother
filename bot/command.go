package bot

type RPCCommand struct {
	Message   string
	Arguments []string
	Code      string
	Host      string
	Nick      string
	Raw       string
	Source    string
	User      string
}
