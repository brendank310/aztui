package resourceviews

import (
	"context"
	"fmt"
	"strings"

	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var resourceGroupSelectItemFuncMap = map[string]func(*ResourceGroupListView, string) tview.Primitive{
	"SpawnResourceTypeListView": (*ResourceGroupListView).SpawnResourceTypeListView,
}

func callResourceGroupMethodByName(view *ResourceGroupListView, methodName string, resourceGroup string) tview.Primitive {
	// Check if the method exists in the map and call it with the receiver
	if method, exists := resourceGroupSelectItemFuncMap[methodName]; exists {
		return method(view, resourceGroup) // Call the method with the receiver
	} else {
		fmt.Printf("Method %s not found\n", methodName)
	}

	return nil
}

type ResourceGroupInfo struct {
	ResourceGroupName string
	ResourceGroupLocation string
	SelectedFunc    func()
}

type ResourceGroupListView struct {
	List           		*tview.List
	StatusBarText  		string
	ActionBarText  		string
	SubscriptionID	 	string
	Parent         		*layout.AppLayout
	ResourceGroupList 	*[]ResourceGroupInfo
}

func NewResourceGroupListView(layout *layout.AppLayout, subscriptionID string) *ResourceGroupListView {
	rg := ResourceGroupListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("Resource Groups (%v)", "F2")

	rg.List.SetBorder(true)
	rg.List.Box.SetTitle(title)
	rg.List.ShowSecondaryText(true)
	rg.ActionBarText = "## Select(Enter) ## | ## Exit(F12) ##"
	rg.SubscriptionID = subscriptionID
	rg.Parent = layout
	layout.FocusedViewIndex = 1

	return &rg
}

func (r *ResourceGroupListView) SpawnResourceTypeListView(resourceGroup string) tview.Primitive {
	// Remove previous views if exist strating from the one at index 2
	r.Parent.RemoveViews(2)

	rtList := NewResourceTypeListView(r.Parent, r.SubscriptionID, resourceGroup)
	rtList.Update()

	return rtList.List
}

func (r *ResourceGroupListView) SelectItem(resourceGroup string) {
	symbolName := GetSymbolName()
	typeName := ExtractTypeName(symbolName)
	fnName := GetFunctionName(symbolName)

	for _, action := range config.GConfig.Actions {
		if typeName == action.Type && fnName == action.Condition {
			p := callResourceGroupMethodByName(r, action.Action, resourceGroup)
			r.Parent.AppendPrimitiveView(p, action.TakeFocus, 1)
		}
	}
}

func (r *ResourceGroupListView) Update() error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}

	r.List.Clear()
	rgClient, err := armresources.NewResourceGroupsClient(r.SubscriptionID, cred, nil)
	if err != nil {
		return fmt.Errorf("failed to create resource groups client: %v", err)
	}

	r.ResourceGroupList = &[]ResourceGroupInfo{}

	rgPager := rgClient.NewListPager(nil)
	for rgPager.More() {
		ctx := context.Background()
		page, err := rgPager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get next resource groups page: %v", err)
		}
		for _, rg := range page.Value {
			resourceGroup := *rg.Name
			location := *rg.Location
			selectedFunc := func() {
				r.SelectItem(resourceGroup)
			}
			*r.ResourceGroupList = append(*r.ResourceGroupList, ResourceGroupInfo{resourceGroup, location, selectedFunc})
			r.List.AddItem(resourceGroup, location, 0, selectedFunc)
		}
	}

	return nil
}

func (r *ResourceGroupListView) UpdateList(layout *layout.AppLayout) error {
	r.List.Clear()
	// Make filtering case insensitive
	filter := strings.ToLower(layout.InputField.GetText())
	for _, ResourceGroupInfo := range *r.ResourceGroupList {
		lowerCaseResourceGroupName := strings.ToLower(ResourceGroupInfo.ResourceGroupName)
		if strings.Contains(lowerCaseResourceGroupName, filter) {
			r.List.AddItem(ResourceGroupInfo.ResourceGroupName, ResourceGroupInfo.ResourceGroupLocation, 0, ResourceGroupInfo.SelectedFunc)
		}
	}
	return nil
}
