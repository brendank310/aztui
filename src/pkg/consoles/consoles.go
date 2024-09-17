package consoles

import (
	"github.com/brendank310/azconsoles/pkg/azconsoles"

	"github.com/gobwas/ws/wsutil"
	"github.com/rivo/tview"
)

func StartSerialConsoleMonitor(subscriptionID string, resourceGroupName string, vmName string, t *tview.TextView) {
	monitorSerial := func(subscriptionID string, resourceGroupName string, vmName string) {
		conn, err := azconsoles.StartSerialConsole(subscriptionID, resourceGroupName, vmName)
		if err != nil {
			panic(err)
		}

		for {
			rxBuf, err := wsutil.ReadServerText(conn)
			if err != nil {
				panic(err)
			}

			t.Write([]byte(tview.TranslateANSI(string(rxBuf))))
		}
	}
	go monitorSerial(subscriptionID, resourceGroupName, vmName)
}
