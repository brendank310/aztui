package consoles

import (
	"github.com/brendank310/azconsoles/pkg/azconsoles"

	"github.com/gobwas/ws/wsutil"
	"github.com/rivo/tview"
)

func StartSerialConsoleMonitor(subscriptionID string, resourceGroupName string, vmName string) *tview.TextView {
	t := tview.NewTextView()
	t.SetTitle(vmName + " Console")
	t.SetBorder(true)
	conn, err := azconsoles.StartSerialConsole(subscriptionID, resourceGroupName, vmName)
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			rxBuf, err := wsutil.ReadServerText(conn)
			if err != nil {
				panic(err)
			}

			t.Write([]byte(tview.TranslateANSI(string(rxBuf))))
		}
	}()

	return t
}
