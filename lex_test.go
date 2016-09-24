package prompt

import "testing"

func TestLexCommand(t *testing.T) {
	testCases := []struct {
		input string
		exp   []item
	}{{"foo", []item{{itemWord, "foo"}}},
		{"foo123", []item{{itemWord, "foo123"}}},
		{"foo123.asd", []item{{itemWord, "foo123.asd"}}},
		{"foo123/asd", []item{{itemWord, "foo123/asd"}}},
		{"foo123%asd", []item{{itemWord, "foo123%asd"}}},
		{"\"test\"", []item{{itemQuotedString, "\"test\""}}},
		{"\"test", []item{{itemError, "unterminated quoted string"}}},
		{"go \"test\"", []item{{itemWord, "go"}, {itemQuotedString, "\"test\""}}},
		{"'test'", []item{{itemQuotedString, "'test'"}}},
		{"`test`", []item{{itemQuotedString, "`test`"}}},
		{"'test", []item{{itemError, "unterminated quoted string"}}},
		{" foo  bar  ",
			[]item{{itemWord, "foo"}, {itemWord, "bar"}}},
		{" foo  \\\nbar  ",
			[]item{{itemWord, "foo"}, {itemLineCont, ""}, {itemWord, "bar"}}},
		{" foo;bar;baz  ",
			[]item{{itemWord, "foo"}, {itemSemi, ";"}, {itemWord, "bar"},
				{itemSemi, ";"}, {itemWord, "baz"}}},
		{"foo|bar",
			[]item{{itemWord, "foo"}, {itemPipe, "|"}, {itemWord, "bar"}}},
		{"  foo |  bar  ",
			[]item{{itemWord, "foo"}, {itemPipe, "|"}, {itemWord, "bar"}}},
		{"  foo = bar",
			[]item{{itemWord, "foo"}, {itemWord, "="}, {itemWord, "bar"}}},
		{"$*:filename",
			[]item{{itemPlaceholder, "*"}, {itemCompletionType, "filename"}}},
		{"  foo $* = bar $*",
			[]item{{itemWord, "foo"}, {itemPlaceholder, "*"}, {itemWord, "="}, {itemWord, "bar"}, {itemPlaceholder, "*"}}},
		{"  foo $ = bar $*",
			[]item{{itemWord, "foo"}, {itemError, "invalid placeholder character ' '"}}},
		{"  foo $* = bar $",
			[]item{{itemWord, "foo"}, {itemPlaceholder, "*"}, {itemWord, "="}, {itemWord, "bar"}, {itemError, "unterminated placeholder"}}},
		{"alias foo $1 $2 = bar $2 $1",
			[]item{{itemWord, "alias"}, {itemWord, "foo"}, {itemPlaceholder, "1"}, {itemPlaceholder, "2"},
				{itemWord, "="}, {itemWord, "bar"}, {itemPlaceholder, "2"}, {itemPlaceholder, "1"}}},
		{"  foo $* = bar $* >a.txt",
			[]item{{itemWord, "foo"}, {itemPlaceholder, "*"}, {itemWord, "="}, {itemWord, "bar"}, {itemPlaceholder, "*"},
				{itemRAngle, ">"}, {itemFilename, "a.txt"}}},
		{"foo $*=bar $*>a.txt",
			[]item{{itemWord, "foo"}, {itemPlaceholder, "*"}, {itemWord, "=bar"}, {itemPlaceholder, "*"},
				{itemRAngle, ">"}, {itemFilename, "a.txt"}}}}

	for _, tc := range testCases {
		exp := tc.exp
		items := lex(tc.input, cmdDescMode)

		i := 0
		for item := range items {
			if i >= len(exp) {
				t.Errorf("got too many items, expected %d (cur = %s)", len(exp), item)
			} else {
				if exp[i].typ != item.typ {
					t.Errorf("expected item %d type = %s, got %s", i, exp[i].typ, item.typ)
				}
				if exp[i].val != item.val {
					t.Errorf("expected item %d val = '%s', got '%s' for test %s", i, exp[i].val, item.val, tc.input)
				}
			}
			i++
		}
		if i != len(tc.exp) {
			t.Errorf("parsed too few items, expected %d, got %d", len(exp), i)
		}

	}

}

func TestLexUserInput(t *testing.T) {
	// same inputs as above, but parse differently in user mode
	testCases := []struct {
		input string
		exp   []item
	}{{"foo", []item{{itemWord, "foo"}}},
		{"foo123", []item{{itemWord, "foo123"}}},
		{"foo123.asd", []item{{itemWord, "foo123.asd"}}},
		{"foo123/asd", []item{{itemWord, "foo123/asd"}}},
		{"foo123%asd", []item{{itemWord, "foo123%asd"}}},
		{"\"test\"", []item{{itemQuotedString, "\"test\""}}},
		{"\"test", []item{{itemError, "unterminated quoted string"}}},
		{"go \"test\"", []item{{itemWord, "go"}, {itemQuotedString, "\"test\""}}},
		{"'test'", []item{{itemQuotedString, "'test'"}}},
		{"`test`", []item{{itemQuotedString, "`test`"}}},
		{"'test", []item{{itemError, "unterminated quoted string"}}},
		{" foo  bar  ",
			[]item{{itemWord, "foo"}, {itemWord, "bar"}}},
		{" foo  \\\nbar  ",
			[]item{{itemWord, "foo"}, {itemLineCont, ""}, {itemWord, "bar"}}},
		{" foo;bar;baz  ",
			[]item{{itemWord, "foo"}, {itemSemi, ";"}, {itemWord, "bar"},
				{itemSemi, ";"}, {itemWord, "baz"}}},
		{"foo|bar",
			[]item{{itemWord, "foo"}, {itemPipe, "|"}, {itemWord, "bar"}}},
		{"  foo |  bar  ",
			[]item{{itemWord, "foo"}, {itemPipe, "|"}, {itemWord, "bar"}}},
		{"  foo = bar",
			[]item{{itemWord, "foo"}, {itemWord, "="}, {itemWord, "bar"}}},
		{"$*:filename", []item{{itemWord, "$*:filename"}}},
		{"  foo $* = bar $*",
			[]item{{itemWord, "foo"}, {itemWord, "$*"}, {itemWord, "="}, {itemWord, "bar"}, {itemWord, "$*"}}},
		{"  foo $ = bar $*",
			[]item{{itemWord, "foo"}, {itemWord, "$"}, {itemWord, "="}, {itemWord, "bar"}, {itemWord, "$*"}}},
		{"  foo $* = bar $",
			[]item{{itemWord, "foo"}, {itemWord, "$*"}, {itemWord, "="}, {itemWord, "bar"}, {itemWord, "$"}}},
		{"alias foo $1 $2 = bar $2 $1",
			[]item{{itemWord, "alias"}, {itemWord, "foo"}, {itemWord, "$1"}, {itemWord, "$2"},
				{itemWord, "="}, {itemWord, "bar"}, {itemWord, "$2"}, {itemWord, "$1"}}},
		{"  foo $* = bar $* >a.txt",
			[]item{{itemWord, "foo"}, {itemWord, "$*"}, {itemWord, "="}, {itemWord, "bar"}, {itemWord, "$*"},
				{itemRAngle, ">"}, {itemFilename, "a.txt"}}},
		{"foo $*=bar $*>a.txt",
			[]item{{itemWord, "foo"}, {itemWord, "$*=bar"}, {itemWord, "$*"},
				{itemRAngle, ">"}, {itemFilename, "a.txt"}}},
	}

	for _, tc := range testCases {
		exp := tc.exp
		items := lex(tc.input, userInputMode)

		i := 0
		for item := range items {
			if i >= len(exp) {
				t.Errorf("got too many items, expected %d (cur = %s)", len(exp), item)
			} else {
				if exp[i].typ != item.typ {
					t.Errorf("expected item %d type = %s, got %s for test %s", i, exp[i].typ, item.typ, tc.input)
				}
				if exp[i].val != item.val {
					t.Errorf("expected item %d val = '%s', got '%s' for test %s", i, exp[i].val, item.val, tc.input)
				}
			}
			i++
		}
		if i != len(tc.exp) {
			t.Errorf("parsed too few items, expected %d, got %d", len(exp), i)
		}

	}

}
