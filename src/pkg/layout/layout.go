package layout

import (
	"fmt"
	"time"

	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/logger"
	"github.com/brendank310/aztui/pkg/utils"
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

	InitKeyBindings[AppLayout, tview.Grid](
		&a, &a, a.Grid, appFuncMap,
	)

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

func (a *AppLayout) Quit() tview.Primitive {
	a.App.Stop()
	return nil
}

type TViewWithSetInputCapture[T any] interface {
	SetInputCapture(capture func(event *tcell.EventKey) *tcell.EventKey) *tview.Box
}

/**
 * InitKeyBindings initializes key bindings for a given layout.
 * The key bindings are based on the configuration file.
 */
func InitKeyBindings[G any, T any](
	layout *AppLayout,
	class *G,
	view TViewWithSetInputCapture[T],
	funcMap map[string]func(*G) tview.Primitive) {
	typeName := utils.GetTypeString[G]()

	// find matching actions
	actions := make([]config.Action, 0)
	for _, view := range config.GConfig.Views {
		if view.Name == typeName {
			actions = view.Actions
			break
		}
	}

	if len(actions) == 0 {
		logger.Println("No actions found for", typeName)
	}

	// find matching key_mappings
	keyActionMap := make(map[config.UserKey]config.Action)
	for _, action := range actions {
		keyActionMap[action.Key] = action
	}

	// set the input capture
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		// check if the key is in the keyActionMap
		logger.Println("Key pressed", event.Key(), event.Rune())
		Ch := rune(0)
		if event.Key() == tcell.KeyRune {
			Ch = event.Rune()
		}

		if action, exists := keyActionMap[config.UserKey{Key: event.Key(), Ch: Ch}]; exists {
			// call the function with the action name
			if method, exists := funcMap[action.Action]; exists {
				logger.Println("Calling method", action.Action)
				view := method(class)
				if view != nil {
					layout.AppendPrimitiveView(view, action.TakeFocus, action.Width)
				}
				return nil
			}
			logger.Println("Method not found", action.Action)
		}

		return event
	})
}
