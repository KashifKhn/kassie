package tui

type View int

const (
	ViewConnection View = iota
	ViewExplorer
	ViewHelp
)

type AppState struct {
	View         View
	PreviousView View
	Width        int
	Height       int
	Profile      string
	Keyspace     string
	Table        string
	Status       string
	Err          error
}

func NewState() AppState {
	return AppState{
		View:         ViewConnection,
		PreviousView: ViewConnection,
		Status:       "Ready",
	}
}

func (s AppState) WithSize(width, height int) AppState {
	s.Width = width
	s.Height = height
	return s
}
