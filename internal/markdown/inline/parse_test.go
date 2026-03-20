package inline

// func TestParse(t *testing.T) {
// 	testCases := []struct {
// 		name    string
// 		input   string
// 		want    []ast.Inline
// 		wantErr error
// 	}{
// 		{
// 			name:    "empty input",
// 			input:   "",
// 			want:    []ast.Inline{},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "single rune",
// 			input: "a",
// 			want: []ast.Inline{
// 				tk.ASTText(),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "plain text",
// 			input: "hello world",
// 			want: []ast.Inline{
// 				tk.ASTText(),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "unicode characters",
// 			input: "café 🎵 — 漢字",
// 			want: []ast.Inline{
// 				tk.ASTText(),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "whitespace only",
// 			input: " \t ",
// 			want: []ast.Inline{
// 				tk.ASTText(),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "unmatched opener",
// 			input: "*abc",
// 			want: []ast.Inline{
// 				tk.ASTText(),
// 				tk.ASTText(),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "unmatched closer",
// 			input: "abc*",
// 			want: []ast.Inline{
// 				tk.ASTText(),
// 				tk.ASTText(),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "literal delimiter in spaced context",
// 			input: "foo * bar",
// 			want: []ast.Inline{
// 				tk.ASTText(),
// 				tk.ASTText(),
// 				tk.ASTText(),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "simple emphasis",
// 			input: "*abc*",
// 			want: []ast.Inline{
// 				tk.ASTEm(tk.ASTText()),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "simple strong",
// 			input: "**abc**",
// 			want: []ast.Inline{
// 				tk.ASTStrong(tk.ASTText()),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "emphasis with surrounding text",
// 			input: "a *b* c",
// 			want: []ast.Inline{
// 				tk.ASTText(),
// 				tk.ASTEm(tk.ASTText()),
// 				tk.ASTText(),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "strong with surrounding text",
// 			input: "a **b** c",
// 			want: []ast.Inline{
// 				tk.ASTText(),
// 				tk.ASTStrong(tk.ASTText()),
// 				tk.ASTText(),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "opener and closer without resolution",
// 			input: "a*b",
// 			want: []ast.Inline{
// 				tk.ASTText(),
// 				tk.ASTText(),
// 				tk.ASTText(),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "simple link",
// 			input: "[x](dest)",
// 			want: []ast.Inline{
// 				tk.ASTLink(tk.ASTText()),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "link with title",
// 			input: `[x](dest "title")`,
// 			want: []ast.Inline{
// 				tk.ASTLink(tk.ASTText()),
// 			},
// 			wantErr: nil,
// 		},
// 	}
//
// 	for _, tc := range testCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			src := source.NewSource(tc.input)
// 			span := source.ByteSpan{
// 				Start: 0,
// 				End:   src.EOF(),
// 			}
//
// 			got, err := Parse(src, span)
//
// 			got = tk.NormalizeASTInlines(got)
// 			want := tk.NormalizeASTInlines(tc.want)
//
// 			assert.Equal(t, got, want)
// 			assert.ErrorIs(t, err, tc.wantErr)
// 		})
// 	}
//
// 	spanCases := []struct {
// 		name    string
// 		input   string
// 		span    *source.ByteSpan
// 		want    []ast.Inline
// 		wantErr error
// 	}{
// 		{
// 			name:  "windowed span yields ast.Text",
// 			input: "prefix: body :suffix",
// 			span:  tk.SpanPtr(8, 12),
// 			want: []ast.Inline{
// 				tk.ASTTextAt(8, 12),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "windowed empty span yields empty",
// 			input:   "hello",
// 			span:    tk.SpanPtr(0, 0),
// 			want:    []ast.Inline{},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "windowed span at beginning",
// 			input: "hello world",
// 			span:  tk.SpanPtr(0, 5),
// 			want: []ast.Inline{
// 				tk.ASTTextAt(0, 5),
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "windowed span at end",
// 			input: "hello world",
// 			span:  tk.SpanPtr(6, 11),
// 			want: []ast.Inline{
// 				tk.ASTTextAt(6, 11),
// 			},
// 			wantErr: nil,
// 		},
// 	}
//
// 	for _, tc := range spanCases {
// 		t.Run(tc.name, func(t *testing.T) {
// 			src := source.NewSource(tc.input)
//
// 			span := source.ByteSpan{
// 				Start: 0,
// 				End:   src.EOF(),
// 			}
// 			if tc.span != nil {
// 				span = *tc.span
// 			}
//
// 			events, err := Scan(src, span)
// 			require.NoError(t, err)
//
// 			got, err := Build(src, span, events)
//
// 			assert.Equal(t, got, tc.want)
// 			assert.ErrorIs(t, err, tc.wantErr)
// 		})
// 	}
// }
