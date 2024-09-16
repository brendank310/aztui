package resourceviews

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/brendank310/aztui/pkg/azcli"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

var virtualMachineSelectItemFuncMap = map[string]func(*VirtualMachineListView, string) tview.Primitive{
	"SpawnVirtualMachineDetailView": (*VirtualMachineListView).SpawnVirtualMachineDetailView,
	"SpawnVirtualMachineCommandListView": (*VirtualMachineListView).SpawnVirtualMachineCommandListView,
}

type VirtualMachineListView struct {
	List           *tview.List
	StatusBarText  string
	ActionBarText  string
	SubscriptionID string
	ResourceGroup  string
	Parent         *layout.AppLayout
}

func NewVirtualMachineListView(layout *layout.AppLayout, subscriptionID string, resourceGroup string) *VirtualMachineListView {
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
	vm.Parent = layout

	return &vm
}

func callVirtualMachineMethodByName(view *VirtualMachineListView, methodName string, vmName string) tview.Primitive {
	if method, exists := virtualMachineSelectItemFuncMap[methodName]; exists {
		return method(view, vmName)
	} else {
		fmt.Printf("Method %s not found\n", methodName)
	}

	return nil
}

func (v *VirtualMachineListView) SelectItem(vmName string) {
	symbolName := GetSymbolName()
	typeName := ExtractTypeName(symbolName)
	fnName := GetFunctionName(symbolName)

	for _, action := range config.GConfig.Actions {
		if typeName == action.Type && fnName == action.Condition {
			p := callVirtualMachineMethodByName(v, action.Action, vmName)
			v.Parent.AppendPrimitiveView(p)
		}
	}
}

func (v *VirtualMachineListView) SpawnVirtualMachineDetailView(vmName string) tview.Primitive {
	t := tview.NewTextView()
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

	t.SetLabel(vmName + " Details")
	text := fmt.Sprintf("Name:\t%v\nResource ID:\t%v\nLocation:\t%v\n",
		*vm.Name,
		*vm.ID,
		*vm.Location)
	t.SetText(text)

	return t
}

func (v *VirtualMachineListView) SpawnVirtualMachineCommandListView(vmName string) tview.Primitive {
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
					cmdForm.AddInputField("Missing argument: " + missingArg, "", 0, nil, func(text string) {
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
						cmdForm.AddInputField("Missing argument: " + arg, "", 0, nil, func(text string) {
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
			output.SetTitle("Command Output")
			output.SetBorder(true)
			output.Write([]byte(out))
		})
	}

	return cmdList
}

func (v *VirtualMachineListView) Update(selectedFunc func()) error {
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
			v.List.AddItem(*vm.Name, "", 0, selectedFunc)
		}
	}

	return nil
}
