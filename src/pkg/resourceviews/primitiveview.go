package resourceviews

import (
	"strings"

	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/logger"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type PrimitiveView interface {
	// Get the name of the view
	Name() string

	// Update the view
	Update() error

	// Set the input capture
	SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey)

	// Get custom input handler. Return nil if no custom input handler is needed
	CustomInputHandler() func(event *tcell.EventKey) *tcell.EventKey

	// Call action based on the action name
	// Return the primitive to add to the layout or error if any
	CallAction(action string) (tview.Primitive, error)

	AppendPrimitiveView(p tview.Primitive, takeFocus bool, width int)

	// SetActionBar
	UpdateActionBar(actionBar *tview.TextView)
}

/**
 * InitKeyBindings initializes key bindings for a given layout.
 * The key bindings are based on the configuration file.
 */
func InitViewKeyBindings(view PrimitiveView) {
	viewName := view.Name()

	// find matching actions
	actions := make([]config.Action, 0)
	for _, view := range config.GConfig.Views {
		if view.Name == viewName {
			actions = view.Actions
			break
		}
	}

	if len(actions) == 0 {
		logger.Println("No actions found for", viewName)
	}

	// find matching key_mappings
	keyActionMap := make(map[string]config.Action)
	for _, action := range actions {
		keyActionMap[action.Key] = action
	}

	// set the input capture
	view.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

		// check if there is a custom input handler
		if view.CustomInputHandler() != nil {
			event = view.CustomInputHandler()(event)
			if event == nil {
				return nil
			}
		}

		// check if the key is in the keyActionMap
		keyName := event.Name()
		keyName = strings.TrimSuffix(strings.TrimPrefix(keyName, "Rune["), "]")
		logger.Println("Key pressed ", keyName)

		if action, exists := keyActionMap[keyName]; exists {
			logger.Println("Action found for key", keyName, action.Action)
			// call the function with the action name
			newView, err := view.CallAction(action.Action)
			if err != nil {
				logger.Println("Error calling action", action.Action, err)
				return event
			}

			if newView != nil {
				view.AppendPrimitiveView(newView, action.TakeFocus, action.Width)
			}
			return nil
		}

		return event
	})
}
