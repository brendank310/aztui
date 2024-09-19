package resourceviews

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var appFuncMap = map[string]func(*AppLayout) tview.Primitive{
	"Quit": (*AppLayout).Quit,
}

type AppLayout struct {
	App       *tview.Application
	Grid      *tview.Grid
	Layout    *tview.Flex
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
			SetRows(1, -6, 1, 1).
			SetBorders(true),
		Layout:    tview.NewFlex(),
		titleBar:  tview.NewTextView().SetLabel("aztui"),
		actionBar: tview.NewTextView().SetLabel("Ctrl-C to exit"),
		statusBar: tview.NewTextView().SetLabel(status),
	}

	a.Grid.AddItem(a.titleBar, 0, 0, 1, 4, 0, 100, false).
		AddItem(a.Layout, 1, 0, 1, 4, 0, 100, true).
		AddItem(a.statusBar, 2, 0, 1, 4, 0, 100, false).
		AddItem(a.actionBar, 3, 0, 1, 4, 0, 100, false)
	a.Layout.SetDirection(tview.FlexColumn)

	InitViewKeyBindings(&a)

	return &a
}

func (a *AppLayout) Name() string {
	return "SubscriptionListView"
}

func (a *AppLayout) SetInputCapture(f func(event *tcell.EventKey) *tcell.EventKey) {
	a.App.SetInputCapture(f)
}

func (a *AppLayout) CustomInputHandler() func(event *tcell.EventKey) *tcell.EventKey {
	return nil
}

func (a *AppLayout) CallAction(action string) (tview.Primitive, error) {
	if actionFunc, ok := appFuncMap[action]; ok {
		return actionFunc(a), nil
	}
	return nil, fmt.Errorf("no action for %s", action)
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

func (a *AppLayout) Update() error {
	return nil
}

func (a *AppLayout) Quit() tview.Primitive {
	a.App.Stop()
	return nil
}
