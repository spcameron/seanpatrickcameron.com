package inline

import (
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/markdown/ast"
	tk "github.com/spcameron/seanpatrickcameron.com/internal/markdown/testkit"
	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

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
			name:  "newlines are treated like ordinary text",
			input: "a\nb",
			want: []Event{
				{Kind: EventText, Lexeme: "a\nb", Position: 0},
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
			name: "one TEXT event emits one ast.Text nodes",
			events: []Event{
				{Kind: EventText, Lexeme: "hello", Position: 0},
			},
			want: []ast.Inline{
				tk.ASTText("hello"),
			},
			wantErr: nil,
		},
		{
			name: "two TEXT events emits two ast.Text nodes",
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
