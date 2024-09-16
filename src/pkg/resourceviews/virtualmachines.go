package resourceviews

import (
	"context"
	"fmt"
	"log"

	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
)

var virtualMachineSelectItemFuncMap = map[string]func(*VirtualMachineListView,string) tview.Primitive {
	"SpawnVirtualMachineDetailView": (*VirtualMachineListView).SpawnVirtualMachineDetailView,
}

type VirtualMachineListView struct {
	List *tview.List
	StatusBarText string
	ActionBarText string
	SubscriptionID string
	ResourceGroup string
	Parent *layout.AppLayout
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

func (v *VirtualMachineListView) SelectItem(vmName string) tview.Primitive {
	symbolName := GetSymbolName()
	typeName := ExtractTypeName(symbolName)
	fnName := GetFunctionName(symbolName)

	for _, action := range config.GConfig.Actions {
		if typeName == action.Type && fnName == action.Condition {
			return callVirtualMachineMethodByName(v, action.Action, vmName)
		}
	}

	return nil
}

func (v *VirtualMachineListView) SpawnVirtualMachineDetailView(vmName string) tview.Primitive {
	return nil
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
