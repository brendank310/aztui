package resourceviews

import (
	"context"
	"fmt"

	"github.com/rivo/tview"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var resourceGroupSelectItemFuncMap = map[string]func(*ResourceGroupListView,string) tview.Primitive {
	"SpawnVirtualMachineListView": (*ResourceGroupListView).SpawnVirtualMachineListView,
}

func callResourceGroupMethodByName(view *ResourceGroupListView, methodName string, resourceGroup string) tview.Primitive {
	// Check if the method exists in the map and call it with the receiver
	if method, exists := resourceGroupSelectItemFuncMap[methodName]; exists {
		return method(view, resourceGroup)  // Call the method with the receiver
	} else {
		fmt.Printf("Method %s not found\n", methodName)
	}

	return nil
}

type ResourceGroupListView struct {
	List *tview.List
	StatusBarText string
	ActionBarText string
	SubscriptionID string
	Parent *layout.AppLayout
}

func NewResourceGroupListView(layout *layout.AppLayout, subscriptionID string) *ResourceGroupListView {
	rg := ResourceGroupListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("Resource Groups (%v)", "F2")

	rg.List.SetBorder(true)
	rg.List.Box.SetTitle(title)
	rg.List.ShowSecondaryText(false)
	rg.ActionBarText = "## Select(Enter) ## | ## Exit(F12) ##"
	rg.SubscriptionID = subscriptionID
	rg.Parent = layout

	return &rg
}

func (r *ResourceGroupListView) SpawnVirtualMachineListView(resourceGroup string) tview.Primitive {
	vmList := NewVirtualMachineListView(r.Parent, r.SubscriptionID, resourceGroup)

	vmList.Update(func() {
		vmName, _ := vmList.List.GetItemText(vmList.List.GetCurrentItem())
		vmList.SelectItem(vmName)
	})

	return vmList.List
}

func (r *ResourceGroupListView) SelectItem(resourceGroup string) {
	symbolName := GetSymbolName()
	typeName := ExtractTypeName(symbolName)
	fnName := GetFunctionName(symbolName)

	for _, action := range config.GConfig.Actions {
		if typeName == action.Type && fnName == action.Condition {
			p := callResourceGroupMethodByName(r, action.Action, resourceGroup)
			r.Parent.AppendPrimitiveView(p)
		}
	}
}

func (r *ResourceGroupListView) Update(selectedFunc func()) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}

	r.List.Clear()
	rgClient, err := armresources.NewResourceGroupsClient(r.SubscriptionID, cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create resource groups client: %v", err)
	}

	rgPager := rgClient.NewListPager(nil)
	for rgPager.More() {
		ctx := context.Background()
		page, err := rgPager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get next resource groups page: %v", err)
		}
		for _, rg := range page.Value {
			name := *rg.Name
			r.List.AddItem(name, "", 0, selectedFunc)
		}
	}

	return nil
}
