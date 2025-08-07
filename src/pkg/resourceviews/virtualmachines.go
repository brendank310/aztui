package resourceviews

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/brendank310/aztui/pkg/azcli"
	"github.com/brendank310/aztui/pkg/cache"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/consoles"
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
	Parent         *AppLayout
}

func NewVirtualMachineListView(appLayout *AppLayout, subscriptionID string, resourceGroup string) *VirtualMachineListView {
	vm := VirtualMachineListView{
		List: tview.NewList(),
	}

	appLayout.FocusedViewIndex = 2

	title := fmt.Sprintf("Virtual Machines (F%v)", appLayout.FocusedViewIndex+1)

	vm.List.SetBorder(true)
	vm.List.Box.SetTitle(title)
	vm.List.ShowSecondaryText(true)
	vm.ActionBarText = ""
	vm.SubscriptionID = subscriptionID
	vm.ResourceGroup = resourceGroup
	vm.Parent = appLayout

	vm.List.SetFocusFunc(func() {
		InitViewKeyBindings(&vm)
		vm.Update()
		vm.UpdateActionBar(vm.Parent.ActionBar)
	})

	return &vm
}

func (v *VirtualMachineListView) UpdateActionBar(t *tview.TextView) {
	actionBarText := ""
	for _, view := range config.GConfig.Views {
		if view.Name == v.Name() {
			for _, action := range view.Actions {
				actionBarText += fmt.Sprintf("%v(%v) | ", action.Description, action.Key)
			}
			actionBarText = actionBarText[:len(actionBarText)-3] // Remove the last " | "
			break
		}
	}

	t.SetText(actionBarText)
}

func (v *VirtualMachineListView) Name() string {
	return "VirtualMachineListView"
}

func (v *VirtualMachineListView) SetInputCapture(f func(event *tcell.EventKey) *tcell.EventKey) {
	v.List.SetInputCapture(f)
}

func (v *VirtualMachineListView) CustomInputHandler() func(event *tcell.EventKey) *tcell.EventKey {
	return nil
}

func (v *VirtualMachineListView) CallAction(action string) (tview.Primitive, error) {
	if actionFunc, ok := virtualMachineSelectItemFuncMap[action]; ok {
		return actionFunc(v), nil
	}
	return nil, fmt.Errorf("no action for %s", action)
}

func (v *VirtualMachineListView) AppendPrimitiveView(p tview.Primitive, takeFocus bool, width int) {
	v.Parent.AppendPrimitiveView(p, takeFocus, width)
}

func (v *VirtualMachineListView) SpawnVirtualMachineDetailView() tview.Primitive {
	vmName, _ := v.List.GetItemText(v.List.GetCurrentItem())
	v.Parent.RemoveViews(4)
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
	v.Parent.RemoveViews(4)
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
				out = fmt.Sprintf("Command execution failed with error: %v\n", err)
			}
			if out == "" {
				out = "Command executed successfully"
			}

			output := tview.NewTextView()
			v.Parent.AppendPrimitiveView(output, false, 0)
			output.SetTitle("Command Output")
			output.SetBorder(true)
			output.Write([]byte(out))
		})
	}

	return cmdList
}

func (v *VirtualMachineListView) Update() error {
	// Use cache service for virtual machine list
	cacheService := GetCacheService()
	if cacheService != nil {
		cacheKey := cache.GenerateVMKey(v.SubscriptionID, v.ResourceGroup)
		
		// Try to get cached VMs first
		data, err := cacheService.GetOrFetch(cacheKey, func() (interface{}, error) {
			return v.fetchVirtualMachines()
		})
		
		if err != nil {
			return err
		}
		
		// Cast the cached data back to the expected type
		if vms, ok := data.([]*armcompute.VirtualMachine); ok {
			v.populateList(vms)
			return nil
		}
	}
	
	// Fallback to direct fetch if cache service is not available
	vms, err := v.fetchVirtualMachines()
	if err != nil {
		return err
	}
	
	v.populateList(vms)
	return nil
}

// fetchVirtualMachines fetches virtual machines from Azure API
func (v *VirtualMachineListView) fetchVirtualMachines() ([]*armcompute.VirtualMachine, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %v", err)
	}

	vmClient, err := armcompute.NewVirtualMachinesClient(v.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create virtual machines client: %v", err)
	}

	var vms []*armcompute.VirtualMachine

	vmPager := vmClient.NewListPager(v.ResourceGroup, nil)
	for vmPager.More() {
		ctx := context.Background()
		page, err := vmPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next virtual machines page: %v", err)
		}

		vms = append(vms, page.Value...)
	}

	return vms, nil
}

// populateList populates the UI list with virtual machine data
func (v *VirtualMachineListView) populateList(vms []*armcompute.VirtualMachine) {
	v.List.Clear()

	if len(vms) == 0 {
		v.List.AddItem("(No VMs in resource group)", "", 0, nil)
		return
	}

	for _, vm := range vms {
		vmName := *vm.Name
		vmLocation := *vm.Location
		v.List.AddItem(vmName, vmLocation, 0, nil)
	}
}
