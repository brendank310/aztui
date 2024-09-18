package resourceviews

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/rivo/tview"
)

var resourceTypeSelectItemFuncMap = map[string]func(*ResourceTypeListView, string) tview.Primitive{
	"SpawnResourceListView": (*ResourceTypeListView).SpawnResourceListView,
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

func (r *ResourceTypeListView) SpawnResourceListView(resourceType string) tview.Primitive {
	// Remove previous views if exist strating from the one at index 1
	r.Parent.RemoveViews(3)

	resourceList := NewResourceListView(r.Parent, r.SubscriptionID, r.ResourceGroup, resourceType)
	resourceList.Update()

	return resourceList.List
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
	// Create a credential using the default Azure credential chain
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("failed to obtain a credential: %v", err)
	}

	// Create a context
	ctx := context.Background()

	// Create a client to interact with the resource management APIs
	resourcesClient, err := armresources.NewClient(r.SubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create resources client: %v", err)
	}

	// Create a pager to list resources in the specified resource group
	pager := resourcesClient.NewListByResourceGroupPager(r.ResourceGroup, nil)

	r.List.Clear()

	// Create a map to store unique resource types
	resourceTypes := make(map[string]struct{})
	// Iterate through the pages and collect resource types
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to get the next page of results: %v", err)
		}

		for _, resource := range page.Value {
			if resource.Type != nil {
				resourceTypes[*resource.Type] = struct{}{}
			}
		}
	}

	for resourceType := range resourceTypes {
		r.List.AddItem(resourceType, "", 0, func() {
			r.SelectItem(resourceType)
		})
	}

	return nil
}
