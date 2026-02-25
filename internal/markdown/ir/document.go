package ir

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

type Document struct {
	Source *source.Source
	Blocks []Block
}
