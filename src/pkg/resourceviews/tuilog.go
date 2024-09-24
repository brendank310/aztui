package resourceviews

import (
       "time"
	"github.com/rivo/tview"

	"github.com/brendank310/aztui/pkg/layout"
	"github.com/brendank310/aztui/pkg/logger"
)

type TUILogView struct {
     View *tview.TextView
     Parent *layout.AppLayout
}

func NewTUILogView(appLayout *layout.AppLayout) *TUILogView {
     tl := TUILogView{
     	    View: tview.NewTextView(),
	    Parent: appLayout,
     }

     tl.View.SetBorder(true)
     tl.View.SetTitle("AzTUI Log")
     go func(t *TUILogView) {
        logs := logger.GetLogs()
     	t.View.Write([]byte(logs))
	logger.ClearLogs()
	time.Sleep(1*time.Second)
     }(&tl)

     return &tl
}
