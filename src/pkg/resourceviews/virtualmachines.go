package resourceviews

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/brendank310/aztui/pkg/azcli"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/consoles"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/brendank310/aztui/pkg/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

var virtualMachineSelectItemFuncMap = map[string]func(*VirtualMachineListView) tview.Primitive{
	"SpawnVirtualMachineDetailView":        (*VirtualMachineListView).SpawnVirtualMachineDetailView,
	"SpawnVirtualMachineSerialConsoleView": (*VirtualMachineListView).SpawnVirtualMachineSerialConsoleView,
	"SpawnVirtualMachineCommandListView":   (*VirtualMachineListView).SpawnVirtualMachineCommandListView,
}

type VirtualMachineListView struct {
	List           *tview.List
	StatusBarText  string
	ActionBarText  string
	SubscriptionID string
	ResourceGroup  string
	Parent         *layout.AppLayout
	FuncMap        map[string]func(*VirtualMachineListView) tview.Primitive
}

func NewVirtualMachineListView(appLayout *layout.AppLayout, subscriptionID string, resourceGroup string) *VirtualMachineListView {
	vm := VirtualMachineListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("Virtual Machines (%v)", "F3")

	vm.List.SetBorder(true)
	vm.List.Box.SetTitle(title)
	vm.List.ShowSecondaryText(false)
	vm.ActionBarText = "## Subscription List(F1) ## | ## Resource Group List(F2) ## | ## Run Command(F5) ## | ## Serial Console (F7) ## | ## Exit(F12) ##"
	vm.SubscriptionID = subscriptionID
	vm.ResourceGroup = resourceGroup
	vm.Parent = appLayout

	for _, action := range config.GConfig.Actions {
		if utils.GetTypeString[VirtualMachineListView]() == action.Type {
			// set the input capture
			vm.List.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

				Ch := rune(0)
				if event.Key() == tcell.KeyRune {
					Ch = event.Rune()
				}
				_ = Ch

				if method, exists := vm.FuncMap[action.Action]; exists {
					view := method(&vm)
					if view != nil {
						vm.Parent.AppendPrimitiveView(view, action.TakeFocus, action.Width)
					}
					return nil
				}
				return event
			})
		}
	}

	vm.Update()
	return &vm
}

func (v *VirtualMachineListView) SpawnVirtualMachineDetailView() tview.Primitive {
	vmName, _ := v.List.GetItemText(v.List.GetCurrentItem())
	t := tview.NewForm()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("Failed to obtain a credential: %v", err)
	}

	// Create a context
	ctx := context.Background()

	// Create a Compute Virtual Machines client
	vmClient, err := armcompute.NewVirtualMachinesClient(v.SubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("Failed to create VM client: %v", err)
	}

	vm, err := vmClient.Get(ctx, v.ResourceGroup, vmName, nil)
	if err != nil {
		log.Fatalf("Failed to get VM: %v", err)
	}

	t.SetTitle(vmName + " Details")
	t.AddInputField("VM Name", *vm.Name, 0, nil, nil).
		AddInputField("Resource ID", *vm.ID, 0, nil, nil).
		AddInputField("Location", *vm.Location, 0, nil, nil).
		AddInputField("OS", string(*vm.Properties.StorageProfile.OSDisk.OSType), 0, nil, nil)
	t.SetBorder(true)

	return t
}

func (v *VirtualMachineListView) SpawnVirtualMachineSerialConsoleView() tview.Primitive {
	vmName, _ := v.List.GetItemText(v.List.GetCurrentItem())
	t := consoles.StartSerialConsoleMonitor(v.SubscriptionID, v.ResourceGroup, vmName)
	t.SetChangedFunc(func() {
		v.Parent.App.Draw()
	})

	return t
}

func (v *VirtualMachineListView) SpawnVirtualMachineCommandListView() tview.Primitive {
	vmName, _ := v.List.GetItemText(v.List.GetCurrentItem())
	cmdMap, err := azcli.GetResourceCommands("vm")
	if err != nil {
		panic(err)
	}

	cmdList := tview.NewList()
	cmdList.SetTitle("VM Commands")
	cmdList.SetBorder(true)
	for k0, v0 := range cmdMap {
		cmdList.AddItem(k0, v0, 0, func() {
			cmdStr, _ := cmdList.GetItemText(cmdList.GetCurrentItem())
			args := []string{"vm", cmdStr, "-g", v.ResourceGroup, "-n", vmName}
			out, err := azcli.RunAzCommand(args, func(a []string, err error) error {
				if strings.HasPrefix(err.Error(), "ERROR: InvalidArgumentValue:") {
					newArgs := a
					cmdForm := tview.NewForm()

					missingArg := strings.Split(err.Error(), "field:")[1]
					cmdForm.AddInputField("Missing argument: "+missingArg, "", 0, nil, func(text string) {
						newArgs = append(newArgs, missingArg)
						newArgs = append(newArgs, text)
						_, _ = azcli.RunAzCommand(newArgs, nil)
					})
				}

				if strings.HasSuffix(err.Error(), "are required\n") {
					newArgs := a
					cmdForm := tview.NewForm()
					extractRequiredArgs := strings.Split(err.Error(), ":")[1]
					missingArgs := strings.Split(strings.TrimSuffix(strings.Replace(strings.Replace(extractRequiredArgs, "(", "", 1), ")", "", 1), " are required\n"), "|")

					for _, arg := range missingArgs {
						cmdForm.AddInputField("Missing argument: "+arg, "", 0, nil, func(text string) {
							newArgs = append(newArgs, arg)
							newArgs = append(newArgs, text)
						})
					}

					_, _ = azcli.RunAzCommand(newArgs, nil)
				}

				return nil
			})
			if err != nil {
				panic(err)
			}

			output := tview.NewTextView()
			v.Parent.AppendPrimitiveView(output, false, 3)
			output.SetTitle("Command Output")
			output.SetBorder(true)
			output.Write([]byte(out))
		})
	}

	return cmdList
}

func (v *VirtualMachineListView) Update() error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}

	v.List.Clear()
	vmClient, err := armcompute.NewVirtualMachinesClient(v.SubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create virtual machines client: %v", err)
	}

	vmPager := vmClient.NewListPager(v.ResourceGroup, nil)
	for vmPager.More() {
		ctx := context.Background()
		page, err := vmPager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to get next virtual machines page: %v", err)
		}

		if len(page.Value) == 0 && !vmPager.More() {
			v.List.AddItem("(No VMs in resource group)", "", 0, nil)
		}

		for _, vm := range page.Value {
			vmName := *vm.Name
			v.List.AddItem(vmName, "", 0, nil)
		}
	}

	return nil
}
