package vm

import (
	"time"
)

type boostrapProps struct {
	Script_Content     string
	Script_Path        string
	Interpreter_Path   string
	Ready_Process_Name string
	Username           string
	Password           string
	Min_Wait_Seconds   int
	Max_Wait_Seconds   int

	//calculated
	Min_Wait time.Duration
	Max_Wait time.Duration
}

type VMProps struct {
	CPU       int
	RAM       int
	Disk      int
	Bootstrap *boostrapProps
}

func (p *VMProps) Initialize() {
	p.Bootstrap.setDurations()
}

func (p VMProps) NeedsBootstrap() bool {
	return p.Bootstrap.Script_Path != "" &&
		p.Bootstrap.Script_Content != "" &&
		p.Bootstrap.Interpreter_Path != "" &&
		p.Bootstrap.Ready_Process_Name != "" &&
		p.Bootstrap.Username != "" &&
		p.Bootstrap.Password != ""
}

func (b *boostrapProps) setDurations() {
	b.Min_Wait = secsIntToDuration(b.Min_Wait_Seconds)
	b.Max_Wait = secsIntToDuration(b.Max_Wait_Seconds)
}

func secsIntToDuration(secs int) time.Duration {
	return time.Duration(float64(secs) * float64(time.Second))
}
