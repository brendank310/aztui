package resourceviews

import (
	"fmt"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var appFuncMap = map[string]func(*AppLayout) tview.Primitive{
	"Quit":            (*AppLayout).Quit,
	"FocusView0":      (*AppLayout).FocusView0,
	"FocusView1":      (*AppLayout).FocusView1,
	"FocusView2":      (*AppLayout).FocusView2,
	"FocusView3":      (*AppLayout).FocusView3,
	"FocusView4":      (*AppLayout).FocusView4,
	"FocusInputField": (*AppLayout).FocusInputField,
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

	a.Grid.AddItem(a.titleBar, 0, 0, 1, 4, 0, 100, false).
		AddItem(a.InputField, 1, 0, 1, 4, 0, 100, true).
		AddItem(a.Layout, 2, 0, 1, 4, 0, 100, false).
		AddItem(a.statusBar, 3, 0, 1, 4, 0, 100, false).
		AddItem(a.ActionBar, 4, 0, 1, 4, 0, 100, false)
	a.Layout.SetDirection(tview.FlexColumn)

	InitViewKeyBindings(&a)

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

func (a *AppLayout) FocusView(index int) {
	if a.Layout.GetItemCount() >= index+1 {
		a.FocusedViewIndex = index
		a.App.SetFocus(a.Layout.GetItem(index))
	}
}

func (a *AppLayout) FocusView0() tview.Primitive {
	a.FocusView(0)
	return nil
}

func (a *AppLayout) FocusView1() tview.Primitive {
	a.FocusView(1)
	return nil
}

func (a *AppLayout) FocusView2() tview.Primitive {
	a.FocusView(2)
	return nil
}

func (a *AppLayout) FocusView3() tview.Primitive {
	a.FocusView(3)
	return nil
}

func (a *AppLayout) FocusView4() tview.Primitive {
	a.FocusView(4)
	return nil
}

func (a *AppLayout) FocusInputField() tview.Primitive {
	a.App.SetFocus(a.InputField)
	return nil
}

func (a *AppLayout) Quit() tview.Primitive {
	a.App.Stop()
	return nil
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

// index : index of a first view that should be removed
func (a *AppLayout) RemoveViews(index int) {
	itemCount := a.Layout.GetItemCount()

	for i := index; i < itemCount; i++ {
		a.Layout.RemoveItem(a.Layout.GetItem(index))
	}
	a.FocusedViewIndex = 0
	a.App.SetFocus(a.Layout.GetItem(0))
}
