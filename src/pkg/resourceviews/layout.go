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
	App              *tview.Application
	Grid             *tview.Grid
	Layout           *tview.Flex
	InputField       *tview.InputField
	titleBar         *tview.TextView
	ActionBar        *tview.TextView
	statusBar        *tview.TextView
	FocusedViewIndex int
}

func NewAppLayout() *AppLayout {
	status := fmt.Sprintf("Status: %v", time.Now().String())

	a := AppLayout{
		App: tview.NewApplication(),
		Grid: tview.NewGrid().
			SetColumns(-1).
			SetRows(1, 1, -6, 1, 1).
			SetBorders(true),
		Layout:           tview.NewFlex(),
		InputField:       tview.NewInputField().SetLabel("(F10) Filter: "),
		titleBar:         tview.NewTextView().SetLabel("aztui"),
		ActionBar:        tview.NewTextView().SetLabel("## Select(Enter) ## | ## Filter(F10) ## | ## Views(F1-F5) ## | ## Exit(Ctrl-C) ##"),
		statusBar:        tview.NewTextView().SetLabel(status),
		FocusedViewIndex: 0,
	}

	a.App.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyF10 {
			a.App.SetFocus(a.InputField)
			return nil
		} else if event.Key() == tcell.KeyF1 {
			if a.Layout.GetItemCount() >= 1 {
				a.FocusedViewIndex = 0
				a.App.SetFocus(a.Layout.GetItem(0))
			}
		} else if event.Key() == tcell.KeyF2 {
			if a.Layout.GetItemCount() >= 2 {
				a.FocusedViewIndex = 1
				a.App.SetFocus(a.Layout.GetItem(1))
			}
		} else if event.Key() == tcell.KeyF3 {
			if a.Layout.GetItemCount() >= 3 {
				a.FocusedViewIndex = 2
				a.App.SetFocus(a.Layout.GetItem(2))
			}
		} else if event.Key() == tcell.KeyF4 {
			if a.Layout.GetItemCount() >= 4 {
				a.FocusedViewIndex = 3
				a.App.SetFocus(a.Layout.GetItem(3))
			}
		} else if event.Key() == tcell.KeyF5 {
			if a.Layout.GetItemCount() >= 5 {
				a.FocusedViewIndex = 4
				a.App.SetFocus(a.Layout.GetItem(4))
			}
		}
		return event
	})

	a.Grid.AddItem(a.titleBar, 0, 0, 1, 4, 0, 100, false).
		AddItem(a.InputField, 1, 0, 1, 4, 0, 100, true).
		AddItem(a.Layout, 2, 0, 1, 4, 0, 100, false).
		AddItem(a.statusBar, 3, 0, 1, 4, 0, 100, false).
		AddItem(a.ActionBar, 4, 0, 1, 4, 0, 100, false)
	a.Layout.SetDirection(tview.FlexColumn)

	// InitViewKeyBindings(&a)

	return &a
}

func (a *AppLayout) Name() string {
	return "AppLayout"
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

// index : index of a first view that should be removed
func (a *AppLayout) RemoveViews(index int) {
	itemCount := a.Layout.GetItemCount()

	for i := index; i < itemCount; i++ {
		a.Layout.RemoveItem(a.Layout.GetItem(index))
	}
	a.FocusedViewIndex = 0
	a.App.SetFocus(a.Layout.GetItem(0))
}
