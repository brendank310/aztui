package main

import (
	_ "bytes"
	"context"
	_ "fmt"
	"log"
	"strings"
	"net"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
	"github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"

	i "github.com/brendank310/beeperland/pkg"
)

var app *tview.Application
var consoleArea *tview.TextView
var wsConn net.Conn

func startConsoleMonitor(subscriptionID string, resourceGroupName string, vmName string) {
	connURL, err := i.StartSerialConsole(subscriptionID, resourceGroupName, vmName)

	_ = connURL

	if err != nil {
		log.Fatalf("Failed to start serial console session %v", err)
	}

	monitorSerial := func(subscriptionID string, resourceGroupName string, vmName string) {
		wsCtx := context.Background()

		connURLSplit := strings.Split(connURL, "?")
		queryParams := strings.Split(connURLSplit[1], "&")

		token := ""
		for _, param := range queryParams {
			if strings.HasPrefix(param, "authorization=") {
				token = strings.Split(param, "=")[1]
			}
		}

		conn, _, _, err := ws.Dial(wsCtx, connURLSplit[0] + "?new=1")
		if err != nil {
			log.Fatalf("Failed to dial websocket %v", err)
		}
		wsConn = conn

		firstMessage := true
		for {
			rxBuf, err := wsutil.ReadServerText(conn)
			if err != nil {
				consoleArea.Write([]byte("failed to read websocket"))
				//log.Fatalf("failed to read websocket: %v", err)
			}

			if firstMessage {
				wsutil.WriteClientText(conn, []byte(token))
				firstMessage = false
				continue
			}
			consoleArea.Write(rxBuf)
		}
	}
	go monitorSerial(subscriptionID, resourceGroupName, vmName)
}

func main() {
	app = tview.NewApplication()

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	ctx := context.Background()

	subClient, err := armsubscriptions.NewClient(cred, nil)
	if err != nil {
		log.Fatalf("failed to create subscriptions client: %v", err)
	}

	subscriptionList := tview.NewList()
	subscriptionList.SetBorder(true)
	subscriptionList.ShowSecondaryText(false)
	subscriptionList.SetWrapAround(true)
	subscriptionList.Box.SetTitle("Subscriptions")

	resourceGroupList := tview.NewList()
	resourceGroupList.SetBorder(true)
	resourceGroupList.ShowSecondaryText(false)
	resourceGroupList.Box.SetTitle("Resource Groups")

	virtualMachineList := tview.NewList()
	virtualMachineList.SetBorder(true)
	virtualMachineList.ShowSecondaryText(false)
	virtualMachineList.Box.SetTitle("Virtual Machines")

	consoleArea = tview.NewTextView()
	consoleArea.SetBorder(true)
	consoleArea.Box.SetTitle("Serial Console")
	consoleArea.SetChangedFunc(func() {
		app.Draw()
	})
	consoleArea.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		wsutil.WriteClientText(wsConn, []byte(string(event.Rune())))
		return nil
	})

	grid := tview.NewGrid().
		SetColumns(-1, -1, -1, -3).
		SetRows(0).
		SetBorders(true)

	virtualMachineList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'r' {
			modal := tview.NewModal().
				SetText("What az vm command would you like to run?").
				AddButtons([]string{"Console", "Update", "Delete", "List Associated IPs", "Restart", "Cancel"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					_, subscriptionID := subscriptionList.GetItemText(subscriptionList.GetCurrentItem())
					resourceGroupName, _ := resourceGroupList.GetItemText(resourceGroupList.GetCurrentItem())
					vmName, _ := virtualMachineList.GetItemText(virtualMachineList.GetCurrentItem())

					if buttonLabel == "Cancel" {
						app.SetRoot(grid, false)
						return
					}

					if buttonLabel == "Console" {
						startConsoleMonitor(subscriptionID, resourceGroupName, vmName)
						app.SetRoot(grid, false)
						app.SetFocus(consoleArea)
						consoleArea.SetScrollable(false)
						return
					}
				})
			app.SetRoot(modal, false)
			return nil
		}
		return event
	})

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	grid.AddItem(     subscriptionList, 0, 0, 0, 0, 0, 50, true).
		AddItem( resourceGroupList, 1, 0, 1, 3, 0, 0, false).
		AddItem(virtualMachineList, 1, 0, 0, 0, 0, 0, false).
		AddItem(       consoleArea, 0, 0, 0, 0, 0, 0, false)

	// Layout for screens wider than 100 cells.
	grid.AddItem(     subscriptionList, 0, 0, 1, 1, 0, 100, true).
		AddItem( resourceGroupList, 0, 1, 1, 1, 0, 100, false).
		AddItem(virtualMachineList, 0, 2, 1, 1, 0, 100, false).
		AddItem(       consoleArea, 0, 3, 1, 1, 0, 100, false)

	// List subscriptions
	subPager := subClient.NewListPager(nil)
	for subPager.More() {
		page, err := subPager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to get next subscriptions page: %v", err)
		}
		for _, subscription := range page.Value {
			subID := *subscription.SubscriptionID
			subName := *subscription.DisplayName
			subscriptionList.AddItem(subName, subID, 0, func() {
				listResourceGroups(ctx, resourceGroupList, virtualMachineList, cred, subID)
				app.SetFocus(resourceGroupList)
			})
		}
	}

	if err := app.SetRoot(grid, true).Run(); err != nil {
		panic(err)
	}
}

func listResourceGroups(ctx context.Context, resourceGroupList *tview.List, virtualMachineList *tview.List, cred *azidentity.DefaultAzureCredential, subscriptionID string) {
	resourceGroupList.Clear()
	rgClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create resource groups client: %v", err)
	}

	rgPager := rgClient.NewListPager(nil)
	for rgPager.More() {
		page, err := rgPager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to get next resource groups page: %v", err)
		}
		for _, rg := range page.Value {
			name := *rg.Name
			resourceGroupList.AddItem(name, "", 0, func() {
				listVirtualMachines(ctx, virtualMachineList, cred, subscriptionID, name)
				app.SetFocus(virtualMachineList)
			})
		}

	}
}

func listVirtualMachines(ctx context.Context, virtualMachineList *tview.List, cred *azidentity.DefaultAzureCredential, subscriptionID string, resourceGroupName string) {
	virtualMachineList.Clear()
	vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create virtual machines client: %v", err)
	}

	vmPager := vmClient.NewListPager(resourceGroupName, nil)
	for vmPager.More() {
		page, err := vmPager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to get next virtual machines page: %v", err)
		}

		if len(page.Value) == 0 && !vmPager.More() {
			virtualMachineList.AddItem("(No VMs in resource group)", "", 0, nil)
		}

		for _, vm := range page.Value {
			virtualMachineList.AddItem(*vm.Name, "", 0, nil)
		}
	}
}
