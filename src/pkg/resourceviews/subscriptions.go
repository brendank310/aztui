package resourceviews

import (
	"context"
	"fmt"
	"strings"

	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

var subscriptionSelectItemFuncMap = map[string]func(*SubscriptionListView, string) tview.Primitive{
	"SpawnResourceGroupListView": (*SubscriptionListView).SpawnResourceGroupListView,
}

type SubscriptionInfo struct {
	SubscriptionName string
	SubscriptionID string
	SelectedFunc func()
}

type SubscriptionListView struct {
	List              *tview.List
	StatusBarText     string
	ActionBarText     string
	Parent            *layout.AppLayout
	SubscriptionList  *[]SubscriptionInfo
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

	s.Update()
	layout.AppendPrimitiveView(s.List, true, 1)
	return &s
}

func (s *SubscriptionListView) SpawnResourceGroupListView(subscriptionID string) tview.Primitive {
	rgList := NewResourceGroupListView(s.Parent, subscriptionID)
	rgList.Update()
	return rgList.List
}

// Function to call a method by name
func callSubscriptionMethodByName(view *SubscriptionListView, methodName string, subscriptionID string) tview.Primitive {
	// Check if the method exists in the map and call it with the receiver
	if method, exists := subscriptionSelectItemFuncMap[methodName]; exists {
		return method(view, subscriptionID) // Call the method with the receiver
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
			s.Parent.AppendPrimitiveView(p, action.TakeFocus, 1)
		}
	}
}

func (s *SubscriptionListView) Update() error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}

	subClient, err := armsubscriptions.NewClient(cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create subscriptions client: %v", err)
	}

	// Initialize the subscription list
	s.SubscriptionList = &[]SubscriptionInfo{}

	// List subscriptions
	subPager := subClient.NewListPager(nil)
	ctx := context.Background()
	for subPager.More() {
		page, err := subPager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get next subscriptions page: %v", err)
		}
		for _, subscription := range page.Value {
			subscriptionID := *subscription.SubscriptionID
			subscriptionName := *subscription.DisplayName
			selectedFunc := func() {
				s.SelectItem(subscriptionID)
			}
			s.List.AddItem(subscriptionName, subscriptionID, 0, selectedFunc)
			*s.SubscriptionList = append(*s.SubscriptionList, SubscriptionInfo{subscriptionName, subscriptionID, selectedFunc})
		}
	}

	return nil
}

func (s *SubscriptionListView) UpdateList(layout *layout.AppLayout) error {
	s.List.Clear()
	// Make filtering case insensitive
	filter := strings.ToLower(layout.InputField.GetText())
	for _,SubscriptionInfo := range *s.SubscriptionList {
		lowerCaseSubscriptionName := strings.ToLower(SubscriptionInfo.SubscriptionName)
		if (strings.Contains(lowerCaseSubscriptionName, filter)) {
			s.List.AddItem(SubscriptionInfo.SubscriptionName, SubscriptionInfo.SubscriptionID, 0, SubscriptionInfo.SelectedFunc)
		}
	}
	return nil
}