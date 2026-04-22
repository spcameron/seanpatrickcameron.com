package ir

import "github.com/spcameron/seanpatrickcameron.com/internal/markdown/source"

// Document is the root IR node produced by block parsing.
type Document struct {
	Source      *source.Source
	Blocks      []Block
	Definitions map[string]ReferenceDefinition
}

// ReferenceDefinition records a parsed link or image reference definition
// in source-oriented form.
type ReferenceDefinition struct {
	FullSpan        source.ByteSpan
	LabelSpan       source.ByteSpan
	DestinationSpan source.ByteSpan
	TitleSpan       source.ByteSpan
	HasTitle        bool
	NormalizedKey   string
}
