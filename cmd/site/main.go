package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/spcameron/seanpatrickcameron.com/templates"
)

func main() {
	log.SetFlags(0)

	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}

	switch os.Args[1] {
	case "build":
		buildCmd(os.Args[2:])
	case "serve":
		serveCmd(os.Args[2:])
	case "-h", "--help", "help":
		usage()
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(2)
	}
}

func usage() {
	usage := `
Usage:
	site build [--out <dir>]
	site serve [--dir <dir>] [--addr <host:port>]

Commands:
	build    Generate static site output (placeholder for now)
	serve    Serve a directory over HTTP for local preview
`

	fmt.Fprintln(os.Stderr, usage)
}

func buildCmd(args []string) {
	fs := flag.NewFlagSet("build", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	out := fs.String("out", "build/public", "output directory")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	if err := os.MkdirAll(*out, 0o755); err != nil {
		log.Fatalf("build: mkdir: %v", err)
	}

	f, err := os.Create(filepath.Join(*out, "index.html"))
	if err != nil {
		log.Fatalf("build: create: %v", err)
	}
	defer f.Close()

	err = templates.Home().Render(context.Background(), f)
	if err != nil {
		log.Fatalf("build: templates render: %v", err)
	}

	log.Printf("build: wrote %s", f.Name)
}

func serveCmd(args []string) {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(os.Stderr)

	dir := fs.String("dir", "build/public", "directory to serve")
	addr := fs.String("addr", "127.0.0.1:8080", "listen address")
	if err := fs.Parse(args); err != nil {
		os.Exit(2)
	}

	info, err := os.Stat(*dir)
	if err != nil {
		log.Fatalf("serve: stat %s: %v", *dir, err)
	}
	if !info.IsDir() {
		log.Fatalf("serve: not a directory: %s", *dir)
	}

	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir(*dir)))

	log.Printf("serve: serving %s on http://%s", *dir, *addr)
	if err := http.ListenAndServe(*addr, mux); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
