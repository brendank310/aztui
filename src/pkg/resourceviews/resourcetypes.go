package resourceviews

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

var resourceTypeSelectItemFuncMap = map[string]func(*ResourceTypeListView) tview.Primitive{
	"SpawnResourceListView": (*ResourceTypeListView).SpawnResourceListView,
}

type ResourceTypeInfo struct {
	Name         string
	ReadableName string
}

type ResourceTypeListView struct {
	List             *tview.List
	StatusBarText    string
	ActionBarText    string
	SubscriptionID   string
	ResourceGroup    string
	Parent           *AppLayout
	ResourceTypeList map[string]ResourceTypeInfo
}

func NewResourceTypeListView(layout *AppLayout, subscriptionID, resourceGroup string) *ResourceTypeListView {
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
	layout.FocusedViewIndex = 2

	InitViewKeyBindings(&rt)

	return &rt
}

func (r *ResourceTypeListView) Name() string {
	return "ResourceTypeListView"
}

func (r *ResourceTypeListView) SetInputCapture(f func(event *tcell.EventKey) *tcell.EventKey) {
	r.List.SetInputCapture(f)
}

func (r *ResourceTypeListView) CustomInputHandler() func(event *tcell.EventKey) *tcell.EventKey {
	return nil
}

func (r *ResourceTypeListView) CallAction(action string) (tview.Primitive, error) {
	if actionFunc, ok := resourceTypeSelectItemFuncMap[action]; ok {
		return actionFunc(r), nil
	}
	return nil, fmt.Errorf("no action for %s", action)
}

func (r *ResourceTypeListView) AppendPrimitiveView(p tview.Primitive, takeFocus bool, width int) {
	r.Parent.AppendPrimitiveView(p, takeFocus, width)
}

func (r *ResourceTypeListView) SpawnResourceListView() tview.Primitive {
	resourceType, _ := r.List.GetItemText(r.List.GetCurrentItem())
	// Remove previous views if exist strating from the one at index 3
	r.Parent.RemoveViews(3)

	resourceList := NewResourceListView(r.Parent, r.SubscriptionID, r.ResourceGroup, resourceType)
	resourceList.Update()

	return resourceList.List
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
	r.ResourceTypeList = make(map[string]ResourceTypeInfo, 0)
	// Iterate through the pages and collect resource types
	for pager.More() {
		page, err := pager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to get the next page of results: %v", err)
		}

		for _, resource := range page.Value {
			if resource.Type != nil {
				resourceType := *resource.Type
				name := resourceType
				readableName := strings.TrimPrefix(resourceType, "Microsoft.")
				(r.ResourceTypeList)[name] = ResourceTypeInfo{name, readableName}
			}
		}
	}

	for _, resourceTypeInfo := range r.ResourceTypeList {
		r.List.AddItem(resourceTypeInfo.ReadableName, "", 0, nil)
	}

	return nil
}
