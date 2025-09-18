package bar

import (
	"fmt"
	"os"

	"github.com/schollz/progressbar/v3"
)

func New(total int, unit, label string) *progressbar.ProgressBar {
	return progressbar.NewOptions(
		total,
		progressbar.OptionSetItsString(unit),
		progressbar.OptionSetDescription(label),
		progressbar.OptionEnableColorCodes(true),
		// default values
		progressbar.OptionSetWidth(10),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() { fmt.Fprint(os.Stderr, "\n") }),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
	)
}
