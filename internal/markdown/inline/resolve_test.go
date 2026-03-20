package inline

// func TestResolve(t *testing.T) {
// 	testCases := []struct {
// 		name    string
// 		input   string
// 		want    CursorSummary
// 		wantErr error
// 	}{
// 		{
// 			name:  "empty input",
// 			input: "",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{},
// 				Delimiters:   []DelimiterSummary{},
// 				Brackets:     []BracketSummary{},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "emphasis",
// 			input: "*abc*",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind:   "node",
// 						Lexeme: "*abc*",
// 						Node: &InlineSummary{
// 							Kind:   "emphasis",
// 							Lexeme: "abc",
// 							Children: []InlineSummary{
// 								{
// 									Kind:   "text",
// 									Lexeme: "abc",
// 								},
// 							},
// 						},
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 0,
// 						CanOpen:      true,
// 						CanClose:     false,
// 						ItemIndex:    0,
// 					},
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 0,
// 						CanOpen:      false,
// 						CanClose:     true,
// 						ItemIndex:    2,
// 					},
// 				},
// 				Brackets: []BracketSummary{},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "strong",
// 			input: "**abc**",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind:   "node",
// 						Lexeme: "**abc**",
// 						Node: &InlineSummary{
// 							Kind:   "strong",
// 							Lexeme: "abc",
// 							Children: []InlineSummary{
// 								{
// 									Kind:   "text",
// 									Lexeme: "abc",
// 								},
// 							},
// 						},
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{
// 					{
// 						Lexeme:       "**",
// 						Delimiter:    '*',
// 						OriginalRun:  2,
// 						RemainingRun: 0,
// 						CanOpen:      true,
// 						CanClose:     false,
// 						ItemIndex:    0,
// 					},
// 					{
// 						Lexeme:       "**",
// 						Delimiter:    '*',
// 						OriginalRun:  2,
// 						RemainingRun: 0,
// 						CanOpen:      false,
// 						CanClose:     true,
// 						ItemIndex:    2,
// 					},
// 				},
// 				Brackets: []BracketSummary{},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "unmatched opener",
// 			input: "*abc",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind:      "delimiter",
// 						Lexeme:    "*",
// 						Delimiter: '*',
// 					},
// 					{
// 						Kind:   "text",
// 						Lexeme: "abc",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 1,
// 						CanOpen:      true,
// 						CanClose:     false,
// 						ItemIndex:    0,
// 					},
// 				},
// 				Brackets: []BracketSummary{},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "unmatched closer",
// 			input: "abc*",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind:   "text",
// 						Lexeme: "abc",
// 					},
// 					{
// 						Kind:      "delimiter",
// 						Lexeme:    "*",
// 						Delimiter: '*',
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 1,
// 						CanOpen:      false,
// 						CanClose:     true,
// 						ItemIndex:    1,
// 					},
// 				},
// 				Brackets: []BracketSummary{},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "neither open nor close",
// 			input: "a * b",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind:   "text",
// 						Lexeme: "a ",
// 					},
// 					{
// 						Kind:      "delimiter",
// 						Lexeme:    "*",
// 						Delimiter: '*',
// 					},
// 					{
// 						Kind:   "text",
// 						Lexeme: " b",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 1,
// 						CanOpen:      false,
// 						CanClose:     false,
// 						ItemIndex:    1,
// 					},
// 				},
// 				Brackets: []BracketSummary{},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "both open and close, without resolution",
// 			input: "a*b",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind:   "text",
// 						Lexeme: "a",
// 					},
// 					{
// 						Kind:      "delimiter",
// 						Lexeme:    "*",
// 						Delimiter: '*',
// 					},
// 					{
// 						Kind:   "text",
// 						Lexeme: "b",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 1,
// 						CanOpen:      true,
// 						CanClose:     true,
// 						ItemIndex:    1,
// 					},
// 				},
// 				Brackets: []BracketSummary{},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "emphasis, mixed surrounding text",
// 			input: "a *b* c",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind:   "text",
// 						Lexeme: "a ",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind:   "node",
// 						Lexeme: "*b*",
// 						Node: &InlineSummary{
// 							Kind:   "emphasis",
// 							Lexeme: "b",
// 							Children: []InlineSummary{
// 								{
// 									Kind:   "text",
// 									Lexeme: "b",
// 								},
// 							},
// 						},
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind:   "text",
// 						Lexeme: " c",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 0,
// 						CanOpen:      true,
// 						CanClose:     false,
// 						ItemIndex:    1,
// 					},
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 0,
// 						CanOpen:      false,
// 						CanClose:     true,
// 						ItemIndex:    3,
// 					},
// 				},
// 				Brackets: []BracketSummary{},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "strong, mixed surrounding text",
// 			input: "a **b** c",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind:   "text",
// 						Lexeme: "a ",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind:   "node",
// 						Lexeme: "**b**",
// 						Node: &InlineSummary{
// 							Kind:   "strong",
// 							Lexeme: "b",
// 							Children: []InlineSummary{
// 								{
// 									Kind:   "text",
// 									Lexeme: "b",
// 								},
// 							},
// 						},
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind:   "text",
// 						Lexeme: " c",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{
// 					{
// 						Lexeme:       "**",
// 						Delimiter:    '*',
// 						OriginalRun:  2,
// 						RemainingRun: 0,
// 						CanOpen:      true,
// 						CanClose:     false,
// 						ItemIndex:    1,
// 					},
// 					{
// 						Lexeme:       "**",
// 						Delimiter:    '*',
// 						OriginalRun:  2,
// 						RemainingRun: 0,
// 						CanOpen:      false,
// 						CanClose:     true,
// 						ItemIndex:    3,
// 					},
// 				},
// 				Brackets: []BracketSummary{},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "triple star run",
// 			input: "***abc***",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind:   "node",
// 						Lexeme: "***abc***",
// 						Node: &InlineSummary{
// 							Kind:   "emphasis",
// 							Lexeme: "abc",
// 							Children: []InlineSummary{
// 								{
// 									Kind:   "strong",
// 									Lexeme: "abc",
// 									Children: []InlineSummary{
// 										{
// 											Kind:   "text",
// 											Lexeme: "abc",
// 										},
// 									},
// 								},
// 							},
// 						},
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{
// 					{
// 						Lexeme:       "***",
// 						Delimiter:    '*',
// 						OriginalRun:  3,
// 						RemainingRun: 0,
// 						CanOpen:      true,
// 						CanClose:     false,
// 						ItemIndex:    0,
// 					},
// 					{
// 						Lexeme:       "***",
// 						Delimiter:    '*',
// 						OriginalRun:  3,
// 						RemainingRun: 0,
// 						CanOpen:      false,
// 						CanClose:     true,
// 						ItemIndex:    2,
// 					},
// 				},
// 				Brackets: []BracketSummary{},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "nested-like interior",
// 			input: "*ab **cd** ef*",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind:   "node",
// 						Lexeme: "*ab **cd** ef*",
// 						Node: &InlineSummary{
// 							Kind:   "emphasis",
// 							Lexeme: "ab **cd** ef",
// 							Children: []InlineSummary{
// 								{
// 									Kind:   "text",
// 									Lexeme: "ab ",
// 								},
// 								{
// 									Kind:   "strong",
// 									Lexeme: "cd",
// 									Children: []InlineSummary{
// 										{
// 											Kind:   "text",
// 											Lexeme: "cd",
// 										},
// 									},
// 								},
// 								{
// 									Kind:   "text",
// 									Lexeme: " ef",
// 								},
// 							},
// 						},
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 0,
// 						CanOpen:      true,
// 						CanClose:     false,
// 						ItemIndex:    0,
// 					},
// 					{
// 						Lexeme:       "**",
// 						Delimiter:    '*',
// 						OriginalRun:  2,
// 						RemainingRun: 0,
// 						CanOpen:      true,
// 						CanClose:     false,
// 						ItemIndex:    2,
// 					},
// 					{
// 						Lexeme:       "**",
// 						Delimiter:    '*',
// 						OriginalRun:  2,
// 						RemainingRun: 0,
// 						CanOpen:      false,
// 						CanClose:     true,
// 						ItemIndex:    4,
// 					},
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 0,
// 						CanOpen:      false,
// 						CanClose:     true,
// 						ItemIndex:    6,
// 					},
// 				},
// 				Brackets: []BracketSummary{},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "simple inline link",
// 			input: `[x](dest)`,
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind:   "node",
// 						Lexeme: `[x](dest)`,
// 						Node: &InlineSummary{
// 							Kind:   "link",
// 							Lexeme: `[x](dest)`,
// 							Children: []InlineSummary{
// 								{
// 									Kind:   "text",
// 									Lexeme: "x",
// 								},
// 							},
// 						},
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{},
// 				Brackets: []BracketSummary{
// 					{
// 						Lexeme:    "]",
// 						ItemIndex: 2,
// 						Active:    false,
// 					},
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "inline link with title",
// 			input: `[x](dest "title")`,
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind:   "node",
// 						Lexeme: `[x](dest "title")`,
// 						Node: &InlineSummary{
// 							Kind:   "link",
// 							Lexeme: `[x](dest "title")`,
// 							Children: []InlineSummary{
// 								{
// 									Kind:   "text",
// 									Lexeme: "x",
// 								},
// 							},
// 						},
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{},
// 				Brackets: []BracketSummary{
// 					{
// 						Lexeme:    "]",
// 						ItemIndex: 2,
// 						Active:    false,
// 					},
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "inline link with empty label",
// 			input: "[](dest)",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind:   "node",
// 						Lexeme: "[](dest)",
// 						Node: &InlineSummary{
// 							Kind:     "link",
// 							Lexeme:   "[](dest)",
// 							Children: []InlineSummary{},
// 						},
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{},
// 				Brackets: []BracketSummary{
// 					{
// 						Lexeme:    "]",
// 						ItemIndex: 1,
// 						Active:    false,
// 					},
// 				},
// 			},
// 			wantErr: nil,
// 		},
// 		{
// 			name:  "inline link with emphasis in label",
// 			input: "[a *b* c](dest)",
// 			want: CursorSummary{
// 				WorkingItems: []WorkingItemSummary{
// 					{
// 						Kind:   "node",
// 						Lexeme: "[a *b* c](dest)",
// 						Node: &InlineSummary{
// 							Kind:   "link",
// 							Lexeme: "[a *b* c](dest)",
// 							Children: []InlineSummary{
// 								{
// 									Kind:   "text",
// 									Lexeme: "a ",
// 								},
// 								{
// 									Kind:   "emphasis",
// 									Lexeme: "b",
// 									Children: []InlineSummary{
// 										{
// 											Kind:   "text",
// 											Lexeme: "b",
// 										},
// 									},
// 								},
// 								{
// 									Kind:   "text",
// 									Lexeme: " c",
// 								},
// 							},
// 						},
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 					{
// 						Kind: "consumed",
// 					},
// 				},
// 				Delimiters: []DelimiterSummary{
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 0,
// 						CanOpen:      true,
// 						CanClose:     false,
// 						ItemIndex:    2,
// 					},
// 					{
// 						Lexeme:       "*",
// 						Delimiter:    '*',
// 						OriginalRun:  1,
// 						RemainingRun: 0,
// 						CanOpen:      false,
// 						CanClose:     true,
// 						ItemIndex:    4,
// 					},
// 				},
// 				Brackets: []BracketSummary{
// 					{
// 						Lexeme:    "]",
// 						ItemIndex: 6,
// 						Active:    false,
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
//
// 			summary := summarizeCursor(cursor)
//
// 			assert.Equal(t, summary, tc.want)
// 			assert.ErrorIs(t, err, tc.wantErr)
// 		})
// 	}
// }
