package output

import (
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

type Progress struct {
	s       *spinner.Spinner
	started time.Time
}

func NewProgress(msg string) *Progress {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Suffix = "  " + msg
	s.Color("cyan")
	s.Start()
	return &Progress{s: s, started: time.Now()}
}

func (p *Progress) Update(msg string) {
	p.s.Suffix = "  " + msg
}

func (p *Progress) Done(msg string) {
	p.s.Stop()
	elapsed := time.Since(p.started).Round(100 * time.Millisecond)
	fmt.Printf("%s %s %s\n",
		SuccessStyle.Render("✓"),
		msg,
		DimStyle.Render(fmt.Sprintf("(%s)", elapsed)),
	)
}

func (p *Progress) Fail(msg string) {
	p.s.Stop()
	fmt.Printf("%s %s\n", ErrorStyle.Render("✗"), msg)
}
