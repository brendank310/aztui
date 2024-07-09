package main

import (
	_ "bytes"
	"context"
	"fmt"
	"log"
	"strings"
	"net"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/runtime"
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

type AzTuiState struct {
	// Basic TUI variables
	app *tview.Application
	grid *tview.Grid
	layout *tview.Flex
	titleBar *tview.TextView
	actionBar *tview.TextView
	statusBar *tview.TextView

	// Top level Azure Lists
	tenantList *tview.List
	subscriptionList *tview.List
	resourceGroupList *tview.List
	virtualMachineList *tview.List

	// Top level Consoles
	cloudShellConsoleView *tview.TextView
	serialConsoleView *tview.TextView

	// Detail Panes
	virtualMachineDetail *tview.TextView

	cred *azidentity.DefaultAzureCredential
	ctx context.Context

	serialConn net.Conn
	serialConnReady bool
}

// consoleArea.SetChangedFunc(func() {
// app.Draw()
// })
func NewAzTuiState() *AzTuiState {
	status := fmt.Sprintf("Status: %v", time.Now().String())
	// Base initialization
	a := AzTuiState{
		app: tview.NewApplication(),
		grid: tview.NewGrid().
			SetColumns(-1, -1, -1, -3).
			SetRows(1, -6, 1, 1).
			SetBorders(true),
		layout: tview.NewFlex(),
		titleBar: tview.NewTextView().SetLabel("Banjo - The AzTUI"),
		actionBar: tview.NewTextView().SetLabel("F12 to exit"),
		statusBar: tview.NewTextView().SetLabel(status),
		tenantList: tview.NewList(),
		subscriptionList: tview.NewList(),
		resourceGroupList: tview.NewList(),
		virtualMachineList: tview.NewList(),
		cloudShellConsoleView: tview.NewTextView(),
		serialConsoleView: tview.NewTextView(),
		virtualMachineDetail: tview.NewTextView(),
		serialConnReady: false,
	}

	a.statusBar.SetChangedFunc(func() {
		a.app.Draw()
	})

	a.serialConsoleView.SetChangedFunc(func() {
		a.app.Draw()
	})

	// Set widget titles
	a.cloudShellConsoleView.Box.SetTitle("Cloud Shell Console")

	a.subscriptionList.Box.SetTitle("Subscriptions (F1)")
	a.resourceGroupList.Box.SetTitle("Resource Groups (F2)")
	a.virtualMachineList.Box.SetTitle("Virtual Machines (F3)")
	a.serialConsoleView.Box.SetTitle("Serial Console (F4)")

	a.tenantList.ShowSecondaryText(false)
	a.subscriptionList.ShowSecondaryText(false).SetBorder(true)
	a.resourceGroupList.ShowSecondaryText(false).SetBorder(true)
	a.virtualMachineList.ShowSecondaryText(false).SetBorder(true)

	a.serialConsoleView.SetBorder(true)
	a.serialConsoleView.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1: {
			a.app.SetFocus(a.subscriptionList)
			return nil
		}
		case tcell.KeyF2: {
			a.app.SetFocus(a.resourceGroupList)
			return nil
		}
		case tcell.KeyF3: {
			a.app.SetFocus(a.virtualMachineList)
			return nil
		}
		default: {
			if a.serialConnReady {
				switch event.Key() {
				case tcell.KeyRune: {
					b := byte(event.Rune())
					wsutil.WriteClientText(a.serialConn, []byte{b})
				}
				default: {
					wsutil.WriteClientText(a.serialConn, []byte(string(event.Rune())))
				}
				}
			}
			return nil
		}
		}
	})

	// Set the input handler for the Virtual Machine List
	a.virtualMachineList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1: {
			a.app.SetFocus(a.subscriptionList)
			return nil
		}
		case tcell.KeyF2: {
			a.app.SetFocus(a.resourceGroupList)
			return nil
		}
		case tcell.KeyF4: {
			a.app.SetFocus(a.serialConsoleView)
			return nil
		}
		case tcell.KeyF5: {
			modal := tview.NewModal().
				SetText("Run az vm command on selected VM:").
				AddButtons([]string{"Console", "Start", "Cancel"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					switch buttonLabel {
					case "Start": {
						_, subscriptionID := a.subscriptionList.GetItemText(a.subscriptionList.GetCurrentItem())
						resourceGroupName, _ := a.resourceGroupList.GetItemText(a.resourceGroupList.GetCurrentItem())
						vmName, _ := a.virtualMachineList.GetItemText(a.virtualMachineList.GetCurrentItem())
						vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, a.cred, nil)
						if err != nil {
							log.Fatalf("failed to create virtual machines client: %v", err)
						}

						startFuture, err := vmClient.BeginStart(a.ctx, resourceGroupName, vmName, nil)
						if err != nil {
							log.Fatalf("failed to start virtual machine: %v", err)
						}

						go func() {
							// Poll until the operation is done
							_, err = startFuture.PollUntilDone(a.ctx, &runtime.PollUntilDoneOptions{
								Frequency: 6 * time.Second,
							})
							if err != nil {
								log.Fatalf("failed to wait for start operation to complete: %v", err)
							}

							a.statusBar.SetLabel(fmt.Sprintf("%v started successfully", vmName))
							a.app.SetRoot(a.grid, false)
							a.app.SetFocus(a.virtualMachineList)
						}()
						a.app.SetRoot(a.grid, false)
						return
					}
					case "Console": {
						a.StartSerialConsoleMonitor()
						a.app.SetRoot(a.grid, false)
						a.app.SetFocus(a.serialConsoleView)
						a.serialConsoleView.SetScrollable(false)
						return
					}
					case "Cancel":
					default: {
						a.app.SetRoot(a.grid, false)
						return
					}
					}
				})
			a.app.SetRoot(modal, false)
			return event
		}
		default: {
		}
		}
		return event
	})

	a.resourceGroupList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF1: {
			a.app.SetFocus(a.subscriptionList)
			return nil
		}
		case tcell.KeyF3: {
			a.app.SetFocus(a.virtualMachineList)
			return nil
		}
		case tcell.KeyF4: {
			a.app.SetFocus(a.serialConsoleView)
			return nil
		}
		default: {
			return event
		}
		}
	})

	a.subscriptionList.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		switch event.Key() {
		case tcell.KeyF2: {
			a.app.SetFocus(a.resourceGroupList)
			return nil
		}
		case tcell.KeyF3: {
			a.app.SetFocus(a.virtualMachineList)
			return nil
		}
		case tcell.KeyF4: {
			a.app.SetFocus(a.serialConsoleView)
			return nil
		}
		case tcell.KeyF12: {
			modal := tview.NewModal().
				SetText("Are you sure you'd like to exit?").
				AddButtons([]string{"Quit", "Cancel"}).
				SetDoneFunc(func(buttonIndex int, buttonLabel string) {
					switch buttonLabel {
					case "Quit": {
						a.app.Stop()
					}
					case "Cancel": {
						a.app.SetRoot(a.grid, true)
					}
					default: {
						a.app.SetRoot(a.grid, true)
					}
					}
				})
			a.app.SetRoot(modal, true)
			return nil
		}
		default: {
			return event
		}
		}
	})

	// Layout for screens narrower than 100 cells (menu and side bar are hidden).
	a.grid.AddItem(     a.subscriptionList, 0, 0, 1, 1, 0, 49, true).
		AddItem( a.resourceGroupList, 0, 1, 1, 1, 0, 0, false).
		AddItem(a.virtualMachineList, 0, 2, 1, 1, 0, 0, false).
		AddItem(       a.serialConsoleView, 0, 3, 1, 1, 0, 0, false)

	// Layout for screens wider than 100 cells.
	a.grid.AddItem(           a.titleBar, 0, 0, 1, 4, 0, 100, false).
		AddItem(  a.subscriptionList, 1, 0, 1, 1, 0, 100, true).
		AddItem( a.resourceGroupList, 1, 1, 1, 1, 0, 100, false).
		AddItem(a.virtualMachineList, 1, 2, 1, 1, 0, 100, false).
		AddItem( a.serialConsoleView, 1, 3, 1, 1, 0, 100, false).
		AddItem(         a.statusBar, 2, 0, 1, 4, 0, 100, false).
		AddItem(         a.actionBar, 3, 0, 1, 4, 0, 100, false)

	a.layout = tview.NewFlex().SetDirection(tview.FlexRow)
	a.layout.AddItem(a.titleBar, 1, 1, false)
	a.layout.AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
		AddItem(a.subscriptionList, 0, 1, true).
		AddItem(a.resourceGroupList, 0, 1, false).
		AddItem(a.virtualMachineList, 0, 1, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(a.virtualMachineDetail, 0, 1, false).
			AddItem(a.serialConsoleView, 0, 1, false), 0, 1, false), 0, 1, false)

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}
	a.cred = cred
	a.ctx = context.Background()

	return &a
}

func (a *AzTuiState) SetErrorStatus(error string) {
	a.statusBar.SetLabel(error)
}

func (a *AzTuiState) StartSerialConsoleMonitor() {
	_, subscriptionID := a.subscriptionList.GetItemText(a.subscriptionList.GetCurrentItem())
	resourceGroupName, _ := a.resourceGroupList.GetItemText(a.resourceGroupList.GetCurrentItem())
	vmName, _ := a.virtualMachineList.GetItemText(a.virtualMachineList.GetCurrentItem())

	connURL, err := i.StartSerialConsole(subscriptionID, resourceGroupName, vmName)

	_ = connURL

	if err != nil {
		a.SetErrorStatus(fmt.Sprintf("Failed to start serial console session %v", err))
		a.app.SetFocus(a.virtualMachineList)
		return
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

		conn, _, _, err := ws.Dial(wsCtx, connURLSplit[0])
		if err != nil {
			a.SetErrorStatus(fmt.Sprintf("failed to dial websocket %v", err))
			a.app.SetFocus(a.virtualMachineList)
			return
		}
		a.serialConn = conn
		a.serialConnReady = true

		firstMessage := true
		for {
			rxBuf, err := wsutil.ReadServerText(conn)
			if err != nil {
				a.SetErrorStatus("failed to read websocket")
				a.app.SetFocus(a.virtualMachineList)
				return
			}

			if firstMessage {
				wsutil.WriteClientText(conn, []byte(token))
				firstMessage = false
			}

			a.serialConsoleView.Write([]byte(tview.TranslateANSI(string(rxBuf))))		}
	}
	go monitorSerial(subscriptionID, resourceGroupName, vmName)
}

// func (a *AzTuiState) UpdateTenantList() {
// 	tenants, err := a.cred.GetTenants(a.ctx, azidentity.TokenCredentialOptions{})
// 	if err != nil {
// 		log.Fatalf("Failed to get tenants: %v", err)
// 	}

// 	// Print the tenant IDs
// 	for _, tenant := range tenants {
// 		a.tenantList.AddItem(tenant.TenantID(), "", 0, func() {
// 			a.UpdateSubscriptionList()
// 			a.app.SetFocus(a.subscriptionList)
// 		})
// 	}
// }

func (a *AzTuiState) UpdateSubscriptionList() {
	subClient, err := armsubscriptions.NewClient(a.cred, nil)
	if err != nil {
		log.Fatalf("failed to create subscriptions client: %v", err)
	}

	a.actionBar.SetLabel("## Select(Enter) ## | ## Exit(F12) ##")

	// List subscriptions
	subPager := subClient.NewListPager(nil)
	for subPager.More() {
		page, err := subPager.NextPage(a.ctx)
		if err != nil {
			log.Fatalf("failed to get next subscriptions page: %v", err)
		}
		for _, subscription := range page.Value {
			subID := *subscription.SubscriptionID
			subName := *subscription.DisplayName
			a.subscriptionList.AddItem(subName, subID, 0, func() {
				a.UpdateResourceGroupList()
				a.app.SetFocus(a.resourceGroupList)
			})
		}
	}
}

func (a *AzTuiState) UpdateResourceGroupList() {
	_, subscriptionID := a.subscriptionList.GetItemText(a.subscriptionList.GetCurrentItem())
	a.actionBar.SetLabel("## Select(Enter) ## | ## Subscription List(F1) ## | ## Exit(F12) ##")
	a.resourceGroupList.Clear()
	rgClient, err := armresources.NewResourceGroupsClient(subscriptionID, a.cred, nil)
	if err != nil {
		log.Fatalf("failed to create resource groups client: %v", err)
	}

	rgPager := rgClient.NewListPager(nil)
	for rgPager.More() {
		page, err := rgPager.NextPage(a.ctx)
		if err != nil {
			log.Fatalf("failed to get next resource groups page: %v", err)
		}
		for _, rg := range page.Value {
			name := *rg.Name
			a.resourceGroupList.AddItem(name, "", 0, func() {
				a.UpdateVirtualMachineList()
				a.app.SetFocus(a.virtualMachineList)
			})
		}
	}
}

func (a *AzTuiState) UpdateVirtualMachineList() {
	_, subscriptionID := a.subscriptionList.GetItemText(a.subscriptionList.GetCurrentItem())
	resourceGroupName, _ := a.resourceGroupList.GetItemText(a.resourceGroupList.GetCurrentItem())

	a.actionBar.SetLabel("## Subscription List(F1) ## | ## Resource Group List(F2) ## | ## Run Command(F5) ## | ## Exit(F12) ##")

	a.virtualMachineList.Clear()
	vmClient, err := armcompute.NewVirtualMachinesClient(subscriptionID, a.cred, nil)
	if err != nil {
		log.Fatalf("failed to create virtual machines client: %v", err)
	}

	vmPager := vmClient.NewListPager(resourceGroupName, nil)
	for vmPager.More() {
		page, err := vmPager.NextPage(a.ctx)
		if err != nil {
			log.Fatalf("failed to get next virtual machines page: %v", err)
		}

		if len(page.Value) == 0 && !vmPager.More() {
			a.virtualMachineList.AddItem("(No VMs in resource group)", "", 0, nil)
		}

		for _, vm := range page.Value {
			a.virtualMachineList.AddItem(*vm.Name, "", 0, nil)
		}
	}
}

func main() {
	a := NewAzTuiState()
	a.UpdateSubscriptionList()

	if err := a.app.SetRoot(a.grid, true).Run(); err != nil {
		panic(err)
	}
}
