package inline

// func TestParseInlineLinkTail(t *testing.T) {
// 	testCases := []struct {
// 		name    string
// 		input   string
// 		want    InlineLinkTail
// 		wantOK  bool
// 		wantErr error
// 	}{
// 		{
// 			name:  "simple destination",
// 			input: `[x](dest)`,
// 			want: InlineLinkTail{
// 				OpenParenItemIndex:  3,
// 				CloseParenItemIndex: 5,
// 				FullSpan:            tk.Span(3, 9),
// 				DestinationSpan:     tk.Span(4, 8),
// 			},
// 			wantOK:  true,
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "leading space before destination",
// 			input: `[x](   dest)`,
// 			want: InlineLinkTail{
// 				OpenParenItemIndex:  3,
// 				CloseParenItemIndex: 5,
// 				FullSpan:            tk.Span(3, 12),
// 				DestinationSpan:     tk.Span(7, 11),
// 			},
// 			wantOK:  true,
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "trailing space after destination",
// 			input: `[x](dest   )`,
// 			want: InlineLinkTail{
// 				OpenParenItemIndex:  3,
// 				CloseParenItemIndex: 5,
// 				FullSpan:            tk.Span(3, 12),
// 				DestinationSpan:     tk.Span(4, 8),
// 			},
// 			wantOK:  true,
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "double quoted title",
// 			input: `[x](dest "title")`,
// 			want: InlineLinkTail{
// 				OpenParenItemIndex:  3,
// 				CloseParenItemIndex: 5,
// 				FullSpan:            tk.Span(3, 17),
// 				DestinationSpan:     tk.Span(4, 8),
// 				TitleSpan:           tk.Span(10, 15),
// 			},
// 			wantOK:  true,
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "single quoted title",
// 			input: `[x](dest 'title')`,
// 			want: InlineLinkTail{
// 				OpenParenItemIndex:  3,
// 				CloseParenItemIndex: 5,
// 				FullSpan:            tk.Span(3, 17),
// 				DestinationSpan:     tk.Span(4, 8),
// 				TitleSpan:           tk.Span(10, 15),
// 			},
// 			wantOK:  true,
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "empty double quoted title",
// 			input: `[x](dest "")`,
// 			want: InlineLinkTail{
// 				OpenParenItemIndex:  3,
// 				CloseParenItemIndex: 5,
// 				FullSpan:            tk.Span(3, 12),
// 				DestinationSpan:     tk.Span(4, 8),
// 				TitleSpan:           tk.Span(10, 10),
// 			},
// 			wantOK:  true,
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "empty single quoted title",
// 			input: `[x](dest '')`,
// 			want: InlineLinkTail{
// 				OpenParenItemIndex:  3,
// 				CloseParenItemIndex: 5,
// 				FullSpan:            tk.Span(3, 12),
// 				DestinationSpan:     tk.Span(4, 8),
// 				TitleSpan:           tk.Span(10, 10),
// 			},
// 			wantOK:  true,
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "extra space before title",
// 			input: `[x](dest    "title")`,
// 			want: InlineLinkTail{
// 				OpenParenItemIndex:  3,
// 				CloseParenItemIndex: 5,
// 				FullSpan:            tk.Span(3, 20),
// 				DestinationSpan:     tk.Span(4, 8),
// 				TitleSpan:           tk.Span(13, 18),
// 			},
// 			wantOK:  true,
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "extra space after title",
// 			input: `[x](dest "title"   )`,
// 			want: InlineLinkTail{
// 				OpenParenItemIndex:  3,
// 				CloseParenItemIndex: 5,
// 				FullSpan:            tk.Span(3, 20),
// 				DestinationSpan:     tk.Span(4, 8),
// 				TitleSpan:           tk.Span(10, 15),
// 			},
// 			wantOK:  true,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "empty destination",
// 			input:   `[x]()`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "whitespace only destination",
// 			input:   `[x](   )`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "double quoted destination",
// 			input:   `[x]("dest")`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "single quoted destination",
// 			input:   `[x]('dest')`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "destintation with parens",
// 			input:   `[x](de(st))`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "missing space before title",
// 			input:   `[x](dest"title")`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "junk after destination",
// 			input:   `[x](dest junk)`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "unterminated double quoted title",
// 			input:   `[x](dest "title)`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "unterminated single quoted title",
// 			input:   `[x](dest 'title)`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "junk after title",
// 			input:   `[x](dest "title" junk)`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "no open paren after bracket",
// 			input:   `[x] dest)`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
// 			wantErr: nil,
// 		},
// 		{
// 			name:    "no closing paren",
// 			input:   `[x](dest`,
// 			want:    InlineLinkTail{},
// 			wantOK:  false,
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
// 			snapshot := slices.Clone(cursor.WorkingItems)
//
// 			require.NotEqual(t, len(cursor.BracketRecords), 0)
// 			closeItemIdx := cursor.BracketRecords[0].ItemIndex
//
// 			require.NotEqual(t, closeItemIdx, -1)
//
// 			tail, ok, err := cursor.tryParseInlineLinkTail(snapshot, closeItemIdx)
//
// 			assert.Equal(t, tail, tc.want)
// 			assert.Equal(t, ok, tc.wantOK)
// 			assert.ErrorIs(t, err, tc.wantErr)
//
// 		})
// 	}
// }
