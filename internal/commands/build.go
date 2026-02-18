package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/spcameron/seanpatrickcameron.com/internal/site"
)

func RunBuild(args []string) ([]string, int, error) {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	out := fs.String("out", "build/public", "output directory")
	if err := fs.Parse(args); err != nil {
		return nil, 2, err
	}

	written, err := site.BuildSite(*out)
	if err != nil {
		return nil, 1, fmt.Errorf("build: %w", err)
	}

	return written, 0, nil
}
