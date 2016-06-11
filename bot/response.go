package bot

type CommandResponse struct {
	Channel string
	Lines   []string
}

func (r *CommandResponse) AppendLine(l string) {
	r.Lines = append(r.Lines, l)
}
