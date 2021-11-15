package ui

type nopUI struct{}

// NewNop creates a nop user interface for testing purposes.
func NewNop() UI {
	return &nopUI{}
}

func (u *nopUI) Printf(string, ...interface{}) {}

func (u *nopUI) GetLevel() Level {
	return None
}

func (u *nopUI) SetLevel(Level) {}

func (u *nopUI) Tracef(Style, string, ...interface{}) {}

func (u *nopUI) Debugf(Style, string, ...interface{}) {}

func (u *nopUI) Infof(Style, string, ...interface{}) {}

func (u *nopUI) Warnf(Style, string, ...interface{}) {}

func (u *nopUI) Errorf(Style, string, ...interface{}) {}

func (u *nopUI) Confrim(string, bool) (bool, error) {
	return true, nil
}

func (u *nopUI) Ask(string, string, ValidateFunc) (string, error) {
	return "", nil
}

func (u *nopUI) AskSecret(string, bool, ValidateFunc) (string, error) {
	return "", nil
}

func (u *nopUI) Select(string, int, []Item, SearchFunc) (Item, error) {
	return Item{}, nil
}
