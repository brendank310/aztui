package resourceviews

import (
	"context"
	"fmt"
	"log"

	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var resourceSelectItemFuncMap = map[string]func(*ResourceListView, string) tview.Primitive{
	"SpawnResourceDetailView": (*ResourceListView).SpawnResourceDetailView,
}

type ResourceListView struct {
	List           *tview.List
	StatusBarText  string
	ActionBarText  string
	SubscriptionID string
	ResourceGroup  string
	ResourceType   string
	Parent         *layout.AppLayout
}

func NewResourceListView(layout *layout.AppLayout, subscriptionID, resourceGroup, resourceType string) *ResourceListView {
	resourceList := ResourceListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("%v (%v)", resourceType, "F4")

	resourceList.List.SetBorder(true)
	resourceList.List.Box.SetTitle(title)
	resourceList.List.ShowSecondaryText(true)
	resourceList.ActionBarText = "## Subscription List(F1) ## | ## Resource Group List(F2) ## | ## Resource Type List(F3) ## | ## Exit(F12) ##"
	resourceList.SubscriptionID = subscriptionID
	resourceList.ResourceGroup = resourceGroup
	resourceList.ResourceType = resourceType
	resourceList.Parent = layout

	return &resourceList
}

func callResourceMethodByName(view *ResourceListView, methodName, resourceName string) tview.Primitive {
	if method, exists := resourceSelectItemFuncMap[methodName]; exists {
		return method(view, resourceName)
	} else {
		fmt.Printf("Method %s not found\n", methodName)
	}

	return nil
}

func (v *ResourceListView) SelectItem(resourceName string) {
	symbolName := GetSymbolName()
	typeName := ExtractTypeName(symbolName)
	fnName := GetFunctionName(symbolName)

	for _, action := range config.GConfig.Actions {
		if typeName == action.Type && fnName == action.Condition {
			p := callResourceMethodByName(v, action.Action, resourceName)
			v.Parent.AppendPrimitiveView(p, action.TakeFocus, 3)
		}
	}
}

func (v *ResourceListView) SpawnResourceDetailView(resourceName string) tview.Primitive {
	// Remove previous views if exist strating from the one at index 4
	v.Parent.RemoveViews(4)

	t := tview.NewForm()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("Failed to obtain a credential: %v", err)
	}

	ctx := context.Background()

	resourcesClient, err := armresources.NewClient(v.SubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create resources client: %v", err)
	}

	filter := fmt.Sprintf("resourceType eq '%s' and name eq '%s'", v.ResourceType, resourceName)

	options := &armresources.ClientListByResourceGroupOptions{
		Filter: &filter,
		Expand: to.Ptr("$expand=createdTime,provisioningState"),
	}

	pager := resourcesClient.NewListByResourceGroupPager(v.ResourceGroup, options)

	var resource *armresources.GenericResourceExpanded
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to get the next page of resources: %v", err)
		}

		if len(page.Value) == 1 {
			resource = page.Value[0]
		} else if len(page.Value) > 1 {
			log.Fatalf("more than one resource found with the name %s", resourceName)
		}
	}

	t.SetTitle(resourceName + " Details")
	t.AddInputField(fmt.Sprintf("%v Name", v.ResourceType), resourceName, 0, nil, nil).
		AddInputField("Resource ID", *resource.ID, 0, nil, nil).
		AddInputField("Location", *resource.Location, 0, nil, nil).
		AddInputField("Provisioning State", *resource.ProvisioningState, 0, nil, nil)
	t.SetBorder(true)

	return t
}

func (v *ResourceListView) Update() error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}

	v.List.Clear()

	ctx := context.Background()

	resourcesClient, err := armresources.NewClient(v.SubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create resources client: %v", err)
	}

	filter := fmt.Sprintf("resourceType eq '%s'", v.ResourceType)

	options := &armresources.ClientListByResourceGroupOptions{
		Filter: &filter,
		Expand: to.Ptr("$expand=createdTime,provisioningState"),
	}

	pager := resourcesClient.NewListByResourceGroupPager(v.ResourceGroup, options)

	if !pager.More() {
		v.List.AddItem(fmt.Sprintf("(No %v in resource group)", v.ResourceType), "", 0, nil)
	}

	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to get the next page of %v: %v", v.ResourceType, err)
		}

		for _, resource := range page.Value {
			v.List.AddItem(*resource.Name, *resource.Location, 0, func() {
				v.SelectItem(*resource.Name)
			})
		}
	}

	return nil
}
