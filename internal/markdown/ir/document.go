package ir

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

type Document struct {
	Source      *source.Source
	Blocks      []Block
	Definitions map[string]ReferenceDefinition
}

type ReferenceDefinition struct {
	FullSpan        source.ByteSpan
	LabelSpan       source.ByteSpan
	DestinationSpan source.ByteSpan
	TitleSpan       source.ByteSpan
	HasTitle        bool
	NormalizedKey   string
}
