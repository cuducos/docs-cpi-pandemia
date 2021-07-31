package bar

import (
	"fmt"
	"os"
	"time"

	"github.com/schollz/progressbar/v3"
)

func New(total int, unit, label string, refresh int) *progressbar.ProgressBar {
	return progressbar.NewOptions(
		total,
		progressbar.OptionSetItsString(unit),
		progressbar.OptionSetDescription(label),
		progressbar.OptionEnableColorCodes(true),
		progressbar.OptionThrottle(time.Duration(refresh)*time.Second),
		// default values
		progressbar.OptionSetWidth(10),
		progressbar.OptionShowCount(),
		progressbar.OptionShowIts(),
		progressbar.OptionOnCompletion(func() { fmt.Fprint(os.Stderr, "\n") }),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
	)
}
