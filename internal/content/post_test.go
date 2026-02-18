package content

import (
	"strings"
	"testing"

	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestSplitPost(t *testing.T) {
	testCases := []struct {
		name    string
		post    []byte
		fm      []byte
		md      []byte
		wantErr error
	}{
		{
			name: "valid post separates into frontmatter and markdown body",
			post: []byte(strings.Join([]string{
				"---",
				"title: Hello",
				"date: 2026-02-17",
				"---",
				"# Hello",
				"",
			}, "\n")),
			fm:      []byte("title: Hello\ndate: 2026-02-17\n"),
			md:      []byte("# Hello\n"),
			wantErr: nil,
		},
		{
			name: "empty frontmatter is structurally allowed",
			post: []byte(strings.Join([]string{
				"---",
				"---",
				"# Hello",
				"",
			}, "\n")),
			fm:      []byte(""),
			md:      []byte("# Hello\n"),
			wantErr: nil,
		},
		{
			name:    "closing fence at EOF (no trailing newline)",
			post:    []byte("---\nkey: val\n---"),
			fm:      []byte("key: val\n"),
			md:      []byte(""),
			wantErr: nil,
		},
		{
			name:    "empty file returns ErrEmptyFile",
			post:    []byte(""),
			wantErr: ErrEmptyFile,
		},
		{
			name:    "missing opening fence returns ErrMissingOpeningFence",
			post:    []byte("# hi\n"),
			wantErr: ErrMissingOpeningFence,
		},
		{
			name:    "opening fence not terminated returns ErrOpeningFenceNotTerminated",
			post:    []byte("---"),
			wantErr: ErrOpeningFenceNotTerminated,
		},
		{
			name: "missing closing fence returns ErrMissingClosingFence",
			post: []byte(strings.Join([]string{
				"---",
				"title: Hello",
				"# Hello",
				"",
			}, "\n")),
			wantErr: ErrMissingClosingFence,
		},
		{
			name:    "opening fence must be exact",
			post:    []byte("--- \n---\n"),
			wantErr: ErrMissingOpeningFence,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fm, md, err := SplitPost(tc.post)

			if tc.wantErr == nil {
				assert.Equal(t, fm, tc.fm)
				assert.Equal(t, md, tc.md)
			} else {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Nil(t, fm)
				assert.Nil(t, md)
			}
		})
	}
}
