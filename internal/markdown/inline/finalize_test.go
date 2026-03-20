package inline

// func TestFinalize(t *testing.T) {
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
// 			name:  "plain text",
// 			input: "hello",
// 			want: []ast.Inline{
// 				ast.Text{Span: tk.Span(0, 5)},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "unmatched opener",
// 			input: "*abc",
// 			want: []ast.Inline{
// 				ast.Text{Span: tk.Span(0, 1)},
// 				ast.Text{Span: tk.Span(1, 4)},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "unmatched closer",
// 			input: "abc*",
// 			want: []ast.Inline{
// 				ast.Text{Span: tk.Span(0, 3)},
// 				ast.Text{Span: tk.Span(3, 4)},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "neither opener nor closer",
// 			input: "foo * bar",
// 			want: []ast.Inline{
// 				ast.Text{Span: tk.Span(0, 4)},
// 				ast.Text{Span: tk.Span(4, 5)},
// 				ast.Text{Span: tk.Span(5, 9)},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "simple emphasis",
// 			input: "*abc*",
// 			want: []ast.Inline{
// 				ast.Em{
// 					Span: tk.Span(1, 4),
// 					Children: []ast.Inline{
// 						ast.Text{Span: tk.Span(1, 4)},
// 					},
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "simple strong",
// 			input: "**abc**",
// 			want: []ast.Inline{
// 				ast.Strong{
// 					Span: tk.Span(2, 5),
// 					Children: []ast.Inline{
// 						ast.Text{Span: tk.Span(2, 5)},
// 					},
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "emphasis with surrounding text",
// 			input: "a *b* c",
// 			want: []ast.Inline{
// 				ast.Text{Span: tk.Span(0, 2)},
// 				ast.Em{
// 					Span: tk.Span(3, 4),
// 					Children: []ast.Inline{
// 						ast.Text{Span: tk.Span(3, 4)},
// 					},
// 				},
// 				ast.Text{Span: tk.Span(5, 7)},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "strong with surrounding text",
// 			input: "a **b** c",
// 			want: []ast.Inline{
// 				ast.Text{Span: tk.Span(0, 2)},
// 				ast.Strong{
// 					Span: tk.Span(4, 5),
// 					Children: []ast.Inline{
// 						ast.Text{Span: tk.Span(4, 5)},
// 					},
// 				},
// 				ast.Text{Span: tk.Span(7, 9)},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "open and close without resolution",
// 			input: "a*b",
// 			want: []ast.Inline{
// 				ast.Text{Span: tk.Span(0, 1)},
// 				ast.Text{Span: tk.Span(1, 2)},
// 				ast.Text{Span: tk.Span(2, 3)},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "simple inline link",
// 			input: `[x](dest)`,
// 			want: []ast.Inline{
// 				ast.Link{
// 					Span:        tk.Span(0, 9),
// 					Label:       tk.Span(1, 2),
// 					Destination: tk.Span(4, 8),
// 					Title:       source.ByteSpan{},
// 					Children: []ast.Inline{
// 						ast.Text{Span: tk.Span(1, 2)},
// 					},
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "inline link with title",
// 			input: `[x](dest "title")`,
// 			want: []ast.Inline{
// 				ast.Link{
// 					Span:        tk.Span(0, 17),
// 					Label:       tk.Span(1, 2),
// 					Destination: tk.Span(4, 8),
// 					Title:       tk.Span(10, 15),
// 					Children: []ast.Inline{
// 						ast.Text{Span: tk.Span(1, 2)},
// 					},
// 				},
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
// 			events, err := Scan(src, span)
// 			require.NoError(t, err)
//
// 			cursor := NewCursor(src, span, events)
//
// 			err = cursor.Gather()
// 			require.NoError(t, err)
//
// 			err = cursor.Resolve()
// 			require.NoError(t, err)
//
// 			inlines, err := cursor.Finalize()
//
// 			assert.Equal(t, inlines, tc.want)
// 			assert.NoError(t, err)
// 		})
// 	}
// }
