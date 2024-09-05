package consoles

import (
	"log"
	"github.com/brendank310/azconsoles/pkg/azconsoles"

	"github.com/gobwas/ws/wsutil"
	"github.com/rivo/tview"
)

func StartSerialConsoleMonitor(subscriptionID string, resourceGroupName string, vmName string, t *tview.TextView) {
	monitorSerial := func(subscriptionID string, resourceGroupName string, vmName string) {
		conn, err := azconsoles.StartSerialConsole(subscriptionID, resourceGroupName, vmName)
		if err != nil {
			log.Printf("%v\n", err)
			return
		}

		for {
			rxBuf, err := wsutil.ReadServerText(conn)
			if err != nil {
				log.Printf("%v\n", err)
				return
			}

			t.Write([]byte(tview.TranslateANSI(string(rxBuf))))
		}
	}
	go monitorSerial(subscriptionID, resourceGroupName, vmName)
}
