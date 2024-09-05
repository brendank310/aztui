package layout

import (
	"fmt"
	"time"

	"github.com/rivo/tview"
)

type AppLayout struct {
	App *tview.Application
	Grid *tview.Grid
	Layout *tview.Flex
	titleBar *tview.TextView
	actionBar *tview.TextView
	statusBar *tview.TextView
}

func NewAppLayout() *AppLayout {
	status := fmt.Sprintf("Status: %v", time.Now().String())

	a := AppLayout{
		App: tview.NewApplication(),
		Grid: tview.NewGrid().
			SetColumns(-1).
			SetRows(1, -6, 1, 1).
			SetBorders(true),
		Layout: tview.NewFlex(),
		titleBar: tview.NewTextView().SetLabel("aztui"),
		actionBar: tview.NewTextView().SetLabel("F12 to exit"),
		statusBar: tview.NewTextView().SetLabel(status),
	}

	a.Grid.AddItem(a.titleBar, 0, 0, 1, 4, 0, 100, false).
		AddItem(a.Layout, 1, 0, 1, 4, 0, 100, true).
		AddItem(a.statusBar, 2, 0, 1, 4, 0, 100, false).
		AddItem(a.actionBar, 3, 0, 1, 4, 0, 100, false)
	a.Layout.SetDirection(tview.FlexColumn)

	return &a
}

func (a *AppLayout) AppendListView(l *tview.List) {
	a.Layout.AddItem(l, 0, 2, true)
	a.App.SetFocus(l)
}

func (a *AppLayout) RemoveListView(l *tview.List) {
	a.Layout.RemoveItem(l)
	a.App.SetFocus(a.Layout)
}

func (a *AppLayout) AppendTextView(t *tview.TextView) {
	a.Layout.AddItem(t, 80, 0, true)
	a.App.SetFocus(t)
}

func (a *AppLayout) RemoveTextView(t *tview.TextView) {
	a.Layout.RemoveItem(t)
	a.App.SetFocus(a.Layout)
}
