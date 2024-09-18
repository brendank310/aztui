package resourceviews

import (
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/utils"	
)

type ResourceView interface {
     CallFunctionByName(name string) tview.Primitive
     AppendInputHandler(f func(event *tcell.EventKey) *tcell.EventKey)
     SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey)
     AppendPrimitiveView(p tview.Primitive, takeFocus bool, width int)
}

// Generic implementation pieces go here
func InstallInputFunctions[R *ResourceView] (r *R) {
	for _, action := range config.GConfig.Actions {
		targetType := utils.GetTypeString[R]()
		if action.Type == targetType {
			a := action.Action
			f := action.TakeFocus
			k := action.Key.Key
			ta := action.Type
			w := action.Width
			r.AppendInputHandler(func(event *tcell.EventKey) *tcell.EventKey {
				if ta == targetType && k == event.Key() {
						view := r.CallFunctionByName(a)
						if view != nil {
						   r.AppendPrimitiveView(view, f, w)
						}
						return nil
					}
					return event
				})
		}
	}

	
}