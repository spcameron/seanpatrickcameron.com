package inline

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestParse(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []ast.Inline
		wantErr error
	}{
		{
			name:  "paragraph text becomes inline sequence",
			input: "a\nb",
			want: []ast.Inline{
				tk.ASTText("a"),
				tk.ASTSoftBreak(),
				tk.ASTText("b"),
			},
			wantErr: nil,
		},
		{
			name:  "hard break survives end-to-end",
			input: "a  \nb",
			want: []ast.Inline{
				tk.ASTText("a"),
				tk.ASTHardBreak(),
				tk.ASTText("b"),
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Parse(tc.input)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestScan(t *testing.T) {
	testCases := []struct {
		name    string
		input   string
		want    []Event
		wantErr error
	}{
		{
			name:    "empty input",
			input:   "",
			want:    []Event{},
			wantErr: nil,
		},
		{
			name:  "single character produces one event",
			input: "a",
			want: []Event{
				{Kind: EventText, Lexeme: "a", Position: 0},
			},
			wantErr: nil,
		},
		{
			name:  "plain text sentence produces one event",
			input: "this is a test",
			want: []Event{
				{Kind: EventText, Lexeme: "this is a test", Position: 0},
			},
			wantErr: nil,
		},
		{
			name:  "unicode round-trips",
			input: "café 🎵",
			want: []Event{
				{Kind: EventText, Lexeme: "café 🎵", Position: 0},
			},
			wantErr: nil,
		},
		{
			name:  "softbreak tokenization",
			input: "a\nb",
			want: []Event{
				{Kind: EventText, Lexeme: "a", Position: 0},
				{Kind: EventSoftBreak, Lexeme: "", Position: 1},
				{Kind: EventText, Lexeme: "b", Position: 2},
			},
			wantErr: nil,
		},
		{
			name:  "hard break tokenization (two spaces)",
			input: "a  \nb",
			want: []Event{
				{Kind: EventText, Lexeme: "a", Position: 0},
				{Kind: EventHardBreak, Lexeme: "", Position: 3},
				{Kind: EventText, Lexeme: "b", Position: 4},
			},
			wantErr: nil,
		},
		{
			name:  "hard break tokenization (backslash)",
			input: "a\\\nb",
			want: []Event{
				{Kind: EventText, Lexeme: "a", Position: 0},
				{Kind: EventHardBreak, Lexeme: "", Position: 2},
				{Kind: EventText, Lexeme: "b", Position: 3},
			},
			wantErr: nil,
		},
		{
			name:  "line consisting of hardbreak only",
			input: "  \n",
			want: []Event{
				{Kind: EventHardBreak, Lexeme: "", Position: 2},
			},
			wantErr: nil,
		},
		{
			name:  "multiple lines",
			input: "a\nb\nc",
			want: []Event{
				{Kind: EventText, Lexeme: "a", Position: 0},
				{Kind: EventSoftBreak, Lexeme: "", Position: 1},
				{Kind: EventText, Lexeme: "b", Position: 2},
				{Kind: EventSoftBreak, Lexeme: "", Position: 3},
				{Kind: EventText, Lexeme: "c", Position: 4},
			},
			wantErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Scan(tc.input)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}

func TestBuild(t *testing.T) {
	testCases := []struct {
		name    string
		events  []Event
		want    []ast.Inline
		wantErr error
	}{
		{
			name:    "empty input",
			events:  []Event{},
			want:    []ast.Inline{},
			wantErr: nil,
		},
		{
			name: "one EventText event emits one ast.Text nodes",
			events: []Event{
				{Kind: EventText, Lexeme: "hello", Position: 0},
			},
			want: []ast.Inline{
				tk.ASTText("hello"),
			},
			wantErr: nil,
		},
		{
			name: "two EventText events emits two ast.Text nodes",
			events: []Event{
				{Kind: EventText, Lexeme: "a", Position: 0},
				{Kind: EventText, Lexeme: "b", Position: 1},
			},
			want: []ast.Inline{
				tk.ASTText("a"),
				tk.ASTText("b"),
			},
			wantErr: nil,
		},
		{
			name: "EventSoftBreak event emits one ast.SoftBreak node",
			events: []Event{
				{Kind: EventSoftBreak, Lexeme: "", Position: 0},
			},
			want: []ast.Inline{
				tk.ASTSoftBreak(),
			},
			wantErr: nil,
		},
		{
			name: "EventHardBreak event emits one ast.HardBreak node",
			events: []Event{
				{Kind: EventHardBreak, Lexeme: "", Position: 0},
			},
			want: []ast.Inline{
				tk.ASTHardBreak(),
			},
			wantErr: nil,
		},
		{
			name: "mixed stream emits corresponding AST nodes",
			events: []Event{
				{Kind: EventText, Lexeme: "a", Position: 0},
				{Kind: EventSoftBreak, Lexeme: "", Position: 1},
				{Kind: EventText, Lexeme: "b", Position: 2},
				{Kind: EventHardBreak, Lexeme: "", Position: 3},
				{Kind: EventText, Lexeme: "c", Position: 4},
			},
			want: []ast.Inline{
				tk.ASTText("a"),
				tk.ASTSoftBreak(),
				tk.ASTText("b"),
				tk.ASTHardBreak(),
				tk.ASTText("c"),
			},
			wantErr: nil,
		},
		{
			name: "unknown event triggers error",
			events: []Event{
				{Kind: EventKind(999), Lexeme: "x", Position: 0},
			},
			want:    nil,
			wantErr: ErrNoRuleMatched,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := Build(tc.events)

			assert.Equal(t, got, tc.want)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
