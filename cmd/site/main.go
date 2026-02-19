package main

import (
	"fmt"
	"log"
	"os"

	"github.com/spcameron/seanpatrickcameron.com/internal/commands"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "build":
		written, code, err := commands.RunBuild(os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(code)
		}

		for _, v := range written {
			fmt.Printf("build: wrote %s\n", v)
		}
	case "serve":
		code, err := commands.RunServe(os.Args[2:])
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			os.Exit(code)
		}
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	const msg = `Usage:
	site build [--out <dir>]
	site serve [--dir <dir>] [--addr <host:port>]

Commands:
	build    Generate static site output (placeholder for now)
	serve    Serve a directory over HTTP for local preview
`
	fmt.Fprint(os.Stderr, msg)
}
