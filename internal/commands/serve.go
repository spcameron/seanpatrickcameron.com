package commands

import (
	"flag"
	"fmt"
	"os"

	"github.com/spcameron/seanpatrickcameron.com/internal/site"
)

func RunServe(args []string) (int, error) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	dir := fs.String("dir", "build/public", "directory to serve")
	addr := fs.String("addr", "127.0.0.1:8080", "listen address")

	if err := fs.Parse(args); err != nil {
		return 2, err
	}

	fmt.Printf("serve: serving %s on http://%s\n", *dir, *addr)

	if err := site.ServeDir(*dir, *addr); err != nil {
		return 1, fmt.Errorf("serve: %w", err)
	}

	return 0, nil
}
