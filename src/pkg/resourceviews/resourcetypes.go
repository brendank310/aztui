package resourceviews

import (
	"fmt"

	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/rivo/tview"
)

var resourceTypeSelectItemFuncMap = map[string]func(*ResourceTypeListView, string) tview.Primitive{
	"SpawnAKSClusterListView":     (*ResourceTypeListView).SpawnAKSClusterListView,
	"SpawnVirtualMachineListView": (*ResourceTypeListView).SpawnVirtualMachineListView,
}

func callResourceTypeMethodByName(view *ResourceTypeListView, methodName string, resourceType string) tview.Primitive {
	// Check if the method exists in the map and call it with the receiver
	if method, exists := resourceTypeSelectItemFuncMap[methodName]; exists {
		return method(view, resourceType) // Call the method with the receiver
	} else {
		fmt.Printf("Method %s not found\n", methodName)
	}

	return nil
}

type ResourceTypeListView struct {
	List           *tview.List
	StatusBarText  string
	ActionBarText  string
	SubscriptionID string
	ResourceGroup  string
	Parent         *layout.AppLayout
}

func NewResourceTypeListView(layout *layout.AppLayout, subscriptionID, resourceGroup string) *ResourceTypeListView {
	rt := ResourceTypeListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("Resource Types (%v)", "F3")

	rt.List.SetBorder(true)
	rt.List.Box.SetTitle(title)
	rt.List.ShowSecondaryText(false)
	rt.ActionBarText = "## Select(Enter) ## | ## Exit(F12) ##"
	rt.SubscriptionID = subscriptionID
	rt.ResourceGroup = resourceGroup
	rt.Parent = layout

	return &rt
}

func (r *ResourceTypeListView) SpawnVirtualMachineListView(resourceType string) tview.Primitive {
	vmList := NewVirtualMachineListView(r.Parent, r.SubscriptionID, r.ResourceGroup)
	vmList.Update()

	return vmList.List
}

func (r *ResourceTypeListView) SpawnAKSClusterListView(resourceType string) tview.Primitive {
	aksList := NewAKSClusterListView(r.Parent, r.SubscriptionID, r.ResourceGroup)
	aksList.Update()

	return aksList.List
}

func (r *ResourceTypeListView) SelectItem(resourceType string) {
	symbolName := GetSymbolName()
	typeName := ExtractTypeName(symbolName)
	fnName := GetFunctionName(symbolName)

	for _, action := range config.GConfig.Actions {
		if typeName == action.Type && fnName == action.Condition {
			p := callResourceTypeMethodByName(r, action.Action, resourceType)
			r.Parent.AppendPrimitiveView(p, action.TakeFocus, 1)
		}
	}
}

func (r *ResourceTypeListView) Update() error {
	r.List.Clear()

	for _, resourceType := range AvailableResourceTypes {
		r.List.AddItem(resourceType, "", 0, func() {
			r.SelectItem(resourceType)
		})
	}

	return nil
}
