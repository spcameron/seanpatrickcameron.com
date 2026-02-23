package inline

type EventKind int

const (
	_ EventKind = iota
	EventText
)

func Scan(input string) ([]Event, error) {
	scanner := NewScanner(input)

	stream := []Event{}
	for {
		event, ok := scanner.Next()
		if !ok {
			break
		}

		stream = append(stream, event)
	}

	return stream, nil
}

type Event struct {
	Kind     EventKind
	Lexeme   string
	Position int
}

type Scanner struct {
	Input    string
	Position int
}

func NewScanner(input string) *Scanner {
	return &Scanner{
		Input:    input,
		Position: 0,
	}
}

func (s *Scanner) EOF() bool {
	return s.Position >= len(s.Input)
}

func (s *Scanner) Next() (Event, bool) {
	if s.EOF() {
		return Event{}, false
	}

	start := s.Position
	kind := EventText
	for s.Position < len(s.Input) {
		_ = s.Input[s.Position]
		s.Position++
	}

	return Event{
		Kind:     kind,
		Lexeme:   s.Input[start:s.Position],
		Position: start,
	}, true
}
