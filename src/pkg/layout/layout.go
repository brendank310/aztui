package layout

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

type AppLayout struct {
	App       *tview.Application
	Grid      *tview.Grid
	Layout    *tview.Flex
	InputField *tview.InputField
	titleBar  *tview.TextView
	actionBar *tview.TextView
	statusBar *tview.TextView
}

func NewAppLayout() *AppLayout {
	status := fmt.Sprintf("Status: %v", time.Now().String())

	a := AppLayout{
		App: tview.NewApplication(),
		Grid: tview.NewGrid().
			SetColumns(-1).
			SetRows(1, 1, -6, 1, 1).
			SetBorders(true),
		Layout:    tview.NewFlex(),
		InputField: tview.NewInputField().SetLabel("Filter here: "),
		titleBar:  tview.NewTextView().SetLabel("aztui"),
		actionBar: tview.NewTextView().SetLabel("Ctrl-C to exit"),
		statusBar: tview.NewTextView().SetLabel(status),
	}

	a.Grid.AddItem(a.titleBar, 0, 0, 1, 4, 0, 100, false).
		AddItem(a.InputField, 1, 0, 1, 4, 0, 100, true).
		AddItem(a.Layout, 2, 0, 1, 4, 0, 100, false).
		AddItem(a.statusBar, 3, 0, 1, 4, 0, 100, false).
		AddItem(a.actionBar, 4, 0, 1, 4, 0, 100, false)
	a.Layout.SetDirection(tview.FlexColumn)

	return &a
}

func (a *AppLayout) AppendPrimitiveView(p tview.Primitive, takeFocus bool, width int) {
	a.Layout.AddItem(p, 0, width, takeFocus)
	if takeFocus {
		a.App.SetFocus(p)
	}
}

func (a *AppLayout) AppendListView(l *tview.List) {
	a.Layout.AddItem(l, 0, 2, true)
}

func (a *AppLayout) RemoveListView(l *tview.List) {
	a.Layout.RemoveItem(l)
	a.App.SetFocus(a.Layout)
}

func (a *AppLayout) AppendTextView(t *tview.TextView) {
	a.Layout.AddItem(t, 80, 0, true)
}

func (a *AppLayout) RemoveTextView(t *tview.TextView) {
	a.Layout.RemoveItem(t)
	a.App.SetFocus(a.Layout)
}
