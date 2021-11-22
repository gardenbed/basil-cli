package release

import "github.com/gardenbed/charm/ui"

const indent = "  "

type changelogUI struct {
	ui.UI
}

func (u *changelogUI) Tracef(style ui.Style, format string, a ...interface{}) {
	u.UI.Tracef(ui.Cyan, indent+format, a...)
}

func (u *changelogUI) Debugf(style ui.Style, format string, a ...interface{}) {
	u.UI.Debugf(ui.Cyan, indent+format, a...)
}

func (u *changelogUI) Infof(style ui.Style, format string, a ...interface{}) {
	u.UI.Infof(ui.Cyan, indent+format, a...)
}

func (u *changelogUI) Warnf(style ui.Style, format string, a ...interface{}) {
	u.UI.Warnf(ui.Yellow, indent+format, a...)
}

func (u *changelogUI) Errorf(style ui.Style, format string, a ...interface{}) {
	u.UI.Warnf(ui.Red, indent+format, a...)
}
