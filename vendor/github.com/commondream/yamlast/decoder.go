package yamlast

import (
	"errors"
	"fmt"
	"strconv"
)

const (
	DocumentNode = 1 << iota
	MappingNode
	SequenceNode
	ScalarNode
	AliasNode
)

// Node represents a node within the AST.
type Node struct {
	Kind         int
	Line, Column int
	Tag          string
	Value        string
	Implicit     bool
	Children     []*Node
	Anchors      map[string]*Node
}

// Parse parses the given bytes and returns the document node for
// the document.
func Parse(b []byte) (*Node, error) {
	parser, err := newParser(b)
	if err != nil {
		return nil, err
	}

	defer parser.destroy()

	return parser.parse()
}

// ----------------------------------------------------------------------------
// Parser, produces a node tree out of a libyaml event stream.

type parser struct {
	parser yaml_parser_t
	event  yaml_event_t
	doc    *Node
}

func newParser(b []byte) (*parser, error) {
	p := parser{}
	if !yaml_parser_initialize(&p.parser) {
		return nil, errors.New("failed to initialize YAML emitter")
	}

	if len(b) == 0 {
		b = []byte{'\n'}
	}

	err := yaml_parser_set_input_string(&p.parser, b)
	if err != nil {
		return nil, err
	}

	err = p.skip()
	if err != nil {
		return nil, err
	}

	if p.event.typ != yaml_STREAM_START_EVENT {
		return nil, errors.New("expected stream start event, got " + strconv.Itoa(int(p.event.typ)))
	}

	err = p.skip()
	if err != nil {
		return nil, err
	}

	return &p, nil
}

func (p *parser) destroy() {
	if p.event.typ != yaml_NO_EVENT {
		yaml_event_delete(&p.event)
	}
	yaml_parser_delete(&p.parser)
}

func (p *parser) skip() error {
	if p.event.typ != yaml_NO_EVENT {
		if p.event.typ == yaml_STREAM_END_EVENT {
			return errors.New("attempted to go past the end of stream; corrupted value?")
		}
		yaml_event_delete(&p.event)
	}
	if !yaml_parser_parse(&p.parser, &p.event) {
		p.fail()
	}

	return nil
}

func (p *parser) fail() error {
	var where string
	var line int
	if p.parser.problem_mark.line != 0 {
		line = p.parser.problem_mark.line
	} else if p.parser.context_mark.line != 0 {
		line = p.parser.context_mark.line
	}
	if line != 0 {
		where = "line " + strconv.Itoa(line) + ": "
	}
	var msg string
	if len(p.parser.problem) > 0 {
		msg = p.parser.problem
	} else {
		msg = "unknown problem parsing YAML content"
	}
	return errors.New(fmt.Sprintf("%s%s", where, msg))
}

func (p *parser) anchor(n *Node, anchor []byte) {
	if anchor != nil {
		p.doc.Anchors[string(anchor)] = n
	}
}

func (p *parser) parse() (*Node, error) {
	switch p.event.typ {
	case yaml_SCALAR_EVENT:
		return p.scalar(), nil
	case yaml_ALIAS_EVENT:
		return p.alias(), nil
	case yaml_MAPPING_START_EVENT:
		return p.mapping()
	case yaml_SEQUENCE_START_EVENT:
		return p.sequence()
	case yaml_DOCUMENT_START_EVENT:
		return p.document()
	case yaml_STREAM_END_EVENT:
		// Happens when attempting to decode an empty buffer.
		return nil, nil
	default:
		return nil, errors.New("attempted to parse unknown event: " + strconv.Itoa(int(p.event.typ)))
	}
}

func (p *parser) node(kind int) *Node {
	return &Node{
		Kind:   kind,
		Line:   p.event.start_mark.line,
		Column: p.event.start_mark.column,
	}
}

func (p *parser) document() (*Node, error) {
	n := p.node(DocumentNode)
	n.Anchors = make(map[string]*Node)
	p.doc = n
	p.skip()
	children, err := p.parse()
	if err != nil {
		return nil, err
	}
	n.Children = append(n.Children, children)
	if p.event.typ != yaml_DOCUMENT_END_EVENT {
		return nil, errors.New("expected end of document event but got " + strconv.Itoa(int(p.event.typ)))
	}
	p.skip()
	return n, nil
}

func (p *parser) alias() *Node {
	n := p.node(AliasNode)
	n.Value = string(p.event.anchor)
	p.skip()
	return n
}

func (p *parser) scalar() *Node {
	n := p.node(ScalarNode)
	n.Value = string(p.event.value)
	n.Tag = string(p.event.tag)
	n.Implicit = p.event.implicit
	p.anchor(n, p.event.anchor)
	p.skip()
	return n
}

func (p *parser) sequence() (*Node, error) {
	n := p.node(SequenceNode)
	p.anchor(n, p.event.anchor)
	p.skip()
	for p.event.typ != yaml_SEQUENCE_END_EVENT {
		children, err := p.parse()
		if err != nil {
			return nil, err
		}
		n.Children = append(n.Children, children)
	}
	p.skip()
	return n, nil
}

func (p *parser) mapping() (*Node, error) {
	n := p.node(MappingNode)
	p.anchor(n, p.event.anchor)
	p.skip()
	for p.event.typ != yaml_MAPPING_END_EVENT {
		key, err := p.parse()
		if err != nil {
			return nil, err
		}
		value, err := p.parse()
		if err != nil {
			return nil, err
		}

		n.Children = append(n.Children, key, value)
	}
	p.skip()
	return n, nil
}
