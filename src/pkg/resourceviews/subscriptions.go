package resourceviews

import (
	"context"
	"fmt"

	"github.com/rivo/tview"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

var subscriptionSelectItemFuncMap = map[string]func(*SubscriptionListView,string) tview.Primitive {
	"SpawnResourceGroupListView": (*SubscriptionListView).SpawnResourceGroupListView,
}

type SubscriptionListView struct {
	List *tview.List
	StatusBarText string
	ActionBarText string
	//ResourceGroupLists []ResourceGroupListView
	Parent *layout.AppLayout
}

func NewSubscriptionListView(layout *layout.AppLayout) *SubscriptionListView {
	s := SubscriptionListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("Subscriptions (%v)", "F1")

	s.List.SetBorder(true)
	s.List.Box.SetTitle(title)
	s.ActionBarText = "## Select(Enter) ## | ## Exit(F12) ##"
	s.Parent = layout

	return &s
}

func (s *SubscriptionListView) SpawnResourceGroupListView(subscriptionID string) tview.Primitive {
	rgList := NewResourceGroupListView(s.Parent, subscriptionID)
	rgList.Update(func() {
		resourceGroup, _ := rgList.List.GetItemText(rgList.List.GetCurrentItem())
		rgList.SelectItem(resourceGroup)
	})

	//s.ResourceGroupLists = append(s.ResourceGroupLists, *rgList)

	//func() {
	//
	// 	vmList := NewVirtualMachineListView(subscriptionID, resourceGroupName)
	// 	a.AppLayout.AppendListView(vmList.List)
	// 	vmList.Update(func() {

	// 		a.AppLayout.AppendListView(cmdList)
	// 	})
	// })
	return rgList.List
}

// Function to call a method by name
func callSubscriptionMethodByName(view *SubscriptionListView, methodName string, subscriptionID string) tview.Primitive {
	// Check if the method exists in the map and call it with the receiver
	if method, exists := subscriptionSelectItemFuncMap[methodName]; exists {
		return method(view, subscriptionID)  // Call the method with the receiver
	} else {
		fmt.Printf("Method %s not found\n", methodName)
	}

	return nil
}

func (s *SubscriptionListView) SelectItem(subscriptionID string) {
	symbolName := GetSymbolName()
	typeName := ExtractTypeName(symbolName)
	fnName := GetFunctionName(symbolName)

	for _, action := range config.GConfig.Actions {
		if typeName == action.Type && fnName == action.Condition {
			p := callSubscriptionMethodByName(s, action.Action, subscriptionID)
			s.Parent.AppendPrimitiveView(p)
		}
	}
}

func (s *SubscriptionListView) Update(selectedFunc func()) error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}

	subClient, err := armsubscriptions.NewClient(cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create subscriptions client: %v", err)
	}

	// List subscriptions
	subPager := subClient.NewListPager(nil)
	ctx := context.Background()
	for subPager.More() {
		page, err := subPager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get next subscriptions page: %v", err)
		}
		for _, subscription := range page.Value {
			subID := *subscription.SubscriptionID
			subName := *subscription.DisplayName
			s.List.AddItem(subName, subID, 0, selectedFunc)
		}
	}

	return nil
}
