package content

import (
	"strings"
	"testing"
	"time"

	"github.com/spcameron/seanpatrickcameron.com/internal/testsupport/assert"
)

func TestDecodeFrontMatter(t *testing.T) {
	testCases := []struct {
		name    string
		data    []byte
		fm      FrontMatter
		wantErr error
	}{
		{
			name: "valid yaml outputs properly formed FrontMatter struct",
			data: []byte(strings.Join([]string{
				"title: test title",
				"date: 1987-06-21",
				"slug: test-slug",
			}, "\n")),
			fm: FrontMatter{
				Title: "test title",
				Date:  time.Date(1987, 06, 21, 0, 0, 0, 0, time.UTC),
				Slug:  "test-slug",
			},
			wantErr: nil,
		},
		{
			name: "invalid yaml fields return ErrInvalidFrontMatter",
			data: []byte(strings.Join([]string{
				"tite: title with typo",
				"date: 1987-06-21",
				"slug: test-slug",
			}, "\n")),
			fm:      FrontMatter{},
			wantErr: ErrInvalidFrontMatter,
		},
		{
			name: "missing title returns ErrMissingTitle",
			data: []byte(strings.Join([]string{
				"date: 1987-06-21",
				"slug: test-slug",
			}, "\n")),
			fm:      FrontMatter{},
			wantErr: ErrMissingTitle,
		},
		{
			name: "missing date returns ErrMissingDate",
			data: []byte(strings.Join([]string{
				"title: test title",
				"slug: test-slug",
			}, "\n")),
			fm:      FrontMatter{},
			wantErr: ErrMissingDate,
		},
		{
			name: "missing slug returns ErrMissingSlug",
			data: []byte(strings.Join([]string{
				"title: test title",
				"date: 1987-06-21",
			}, "\n")),
			fm:      FrontMatter{},
			wantErr: ErrMissingSlug,
		},
		{
			name: "invalid date returns ErrInvalidDate",
			data: []byte(strings.Join([]string{
				"title: test title",
				"date: invalid date",
				"slug: test-slug",
			}, "\n")),
			fm:      FrontMatter{},
			wantErr: ErrInvalidDate,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fm, err := DecodeFrontMatter(tc.data)

			assert.Equal(t, fm, tc.fm)
			assert.ErrorIs(t, err, tc.wantErr)
		})
	}
}
