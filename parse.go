package prompt

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
)

type input struct {
	words      []segment
	filters    []filter
	outputFile string
}

func (i input) asUser() string {
	b := bytes.Buffer{}
	// command string
	for j, w := range i.words {
		b.WriteString(w.value)
		if j+1 != len(i.words) {
			b.WriteRune(' ')
		}
	}

	// filters
	if len(i.filters) > 0 {
		for j, f := range i.filters {
			b.WriteString(" | ")
			b.WriteString(f.cmd)
			if len(f.args) > 0 {
				b.WriteRune(' ')
				b.WriteString(strings.Join(f.args, " "))
			}
			if j+1 != len(i.filters) {
				b.WriteRune(' ')
			}
		}
	}

	// redirection
	if i.outputFile != "" {
		b.WriteString(" > ")
		b.WriteString(i.outputFile)
	}

	return b.String()
}

//go:generate stringer -type=segmentType
type segmentType byte

const (
	wordType segmentType = iota
	placeholderType
)

type segment struct {
	value string
	typ   segmentType
	ctype string // completion type
}

type filter struct {
	cmd  string
	args []string
}

func (s segment) String() string {
	b := bytes.Buffer{}
	b.WriteRune('<')
	b.WriteString(s.typ.String())
	b.WriteRune(' ')
	b.WriteString(s.value)
	if s.ctype != "" {
		b.WriteString(" complete:")
		b.WriteString(s.ctype)
	}
	b.WriteRune('>')
	return b.String()
}

func (i input) String() string {
	b := bytes.Buffer{}
	b.WriteRune('{')

	for wc, w := range i.words {
		b.WriteString(w.String())
		if wc+1 != len(i.words) {
			b.WriteRune(' ')
		}
	}

	if len(i.filters) > 0 {
		b.WriteString(" filter: ")
		for fc, f := range i.filters {
			b.WriteString(f.cmd)
			b.WriteString(fmt.Sprintf("%v", f.args))
			if fc+1 != len(i.filters) {
				b.WriteRune(' ')
			}
		}
	}

	b.WriteRune('}')
	return b.String()
}

type parser struct {
	items      []item
	lexedItems chan item
	curInput   input
	parsed     []input
	err        error
}

func (p *parser) next() item {
	if len(p.items) > 0 {
		r := p.items[0]
		p.items = p.items[1:]
		return r
	}
	return <-p.lexedItems
}

func (p *parser) peek() item {
	if len(p.items) > 0 {
		return p.items[0]
	}

	p.items = append(p.items, <-p.lexedItems)
	return p.items[0]
}

func (p *parser) backup(i item) {
	p.items = append(p.items, i)
}

type parseStateFn func(*parser) parseStateFn

func parseOutputfile(p *parser) parseStateFn {
	file := p.next()
	if file.typ != itemFilename {
		p.err = errors.New("expected output filename")
		return nil
	}
	if p.curInput.outputFile != "" {
		p.err = errors.New("cannot specify multiple output files")
		return nil
	}
	p.curInput.outputFile = file.val

	nItem := p.next()
	switch nItem.typ {
	case itemEOF:
		fallthrough
	case itemChanClose:
		return nil
	case itemSemi:
		return parseStartCmd
	default:
		p.err = fmt.Errorf("unexpected %s token '%s'", nItem.typ, nItem.val)
		return nil
	}
}

func parseFilter(p *parser) parseStateFn {
	filterCmd := p.next()
	if filterCmd.typ != itemWord {
		p.err = errors.New("expected word after |")
		return nil
	}
	f := filter{}
	f.cmd = filterCmd.val
	for fa := p.peek(); fa.typ == itemWord; {
		f.args = append(f.args, fa.val)
		p.next()
		fa = p.peek()
	}

	p.curInput.filters = append(p.curInput.filters, f)
	return parseMidCmd
}

func parsePlaceholder(p *parser) parseStateFn {
	item := p.next()
	seg := segment{item.val, placeholderType, ""}
	if p.peek().typ == itemCompletionType {
		seg.ctype = p.next().val
	}

	for _, ow := range p.curInput.words {
		if ow.typ == placeholderType && ow.value == seg.value {
			p.err = fmt.Errorf("duplicate placeholder $%s", seg.value)
			return nil
		}
	}
	p.curInput.words = append(p.curInput.words, seg)
	return parseMidCmd
}

// starting a new commmand
func parseStartCmd(p *parser) parseStateFn {
	if len(p.curInput.words) > 0 {
		p.parsed = append(p.parsed, p.curInput)
	}

	p.curInput = input{}
	for {
		item := p.next()
		switch item.typ {
		case itemEOF:
			fallthrough
		case itemChanClose:
			return nil
		case itemWord:
			p.backup(item)
			return parseMidCmd
		default:
			p.err = fmt.Errorf("unexpected %s token '%s'", item.typ, item.val)
			return nil
		}
	}
}

// state once we know we're parsing a command
func parseMidCmd(p *parser) parseStateFn {
	for {
		item := p.next()
		switch item.typ {
		case itemEOF:
			fallthrough
		case itemChanClose:
			return nil
		case itemError:
			p.err = errors.New(item.val)
			return nil
		case itemWord:
			p.curInput.words = append(p.curInput.words, segment{item.val, wordType, ""})
		// $*, $1, $2, etc.
		case itemPlaceholder:
			p.backup(item)
			return parsePlaceholder
		// |
		case itemPipe:
			return parseFilter
		// >
		case itemRAngle:
			return parseOutputfile
		// ;
		case itemSemi:
			return parseStartCmd
		default:
			p.err = fmt.Errorf("unexpected %s token '%s'", item.typ, item.val)
			return nil
		}
	}
}

func parse(userInput string) ([]input, error) {
	p := &parser{lexedItems: lex(userInput)}
	return p.run()
}

func (p *parser) run() ([]input, error) {
	for state := parseStartCmd; state != nil; {
		state = state(p)
	}
	if len(p.curInput.words) > 0 {
		p.parsed = append(p.parsed, p.curInput)
	}
	if p.err != nil {
		return nil, p.err
	}
	return p.parsed, nil
}
