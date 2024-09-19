package resourceviews

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/brendank310/aztui/pkg/logger"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var resourceSelectItemFuncMap = map[string]func(*ResourceListView) tview.Primitive{
	"SpawnResourceDetailView": (*ResourceListView).SpawnResourceDetailView,
}

type ResourceListView struct {
	List           *tview.List
	StatusBarText  string
	ActionBarText  string
	SubscriptionID string
	ResourceGroup  string
	ResourceType   string
	ReadableName   string
	Parent         *AppLayout
}

func NewResourceListView(layout *AppLayout, subscriptionID, resourceGroup, resourceType string) *ResourceListView {
	resourceList := ResourceListView{
		List: tview.NewList(),
	}

	resourceList.ReadableName = strings.TrimPrefix(resourceType, "Microsoft.")

	title := fmt.Sprintf("%v (%v)", resourceList.ReadableName, "F4")

	resourceList.List.SetBorder(true)
	resourceList.List.Box.SetTitle(title)
	resourceList.List.ShowSecondaryText(true)
	resourceList.ActionBarText = "## Subscription List(F1) ## | ## Resource Group List(F2) ## | ## Resource Type List(F3) ## | ## Exit(F12) ##"
	resourceList.SubscriptionID = subscriptionID
	resourceList.ResourceGroup = resourceGroup
	resourceList.ResourceType = resourceType
	resourceList.Parent = layout
	layout.FocusedViewIndex = 3

	InitViewKeyBindings(&resourceList)

	resourceList.Update()

	return &resourceList
}

func (v *ResourceListView) Name() string {
	return "ResourceListView"
}

func (v *ResourceListView) SetInputCapture(f func(event *tcell.EventKey) *tcell.EventKey) {
	v.List.SetInputCapture(f)
}

func (v *ResourceListView) CustomInputHandler() func(event *tcell.EventKey) *tcell.EventKey {
	return nil
}

func (v *ResourceListView) CallAction(action string) (tview.Primitive, error) {
	if actionFunc, ok := resourceSelectItemFuncMap[action]; ok {
		return actionFunc(v), nil
	}
	return nil, fmt.Errorf("no action for %s", action)
}

func (v *ResourceListView) AppendPrimitiveView(p tview.Primitive, takeFocus bool, width int) {
	v.Parent.AppendPrimitiveView(p, takeFocus, width)
}

func (v *ResourceListView) SpawnResourceDetailView() tview.Primitive {
	resourceName, _ := v.List.GetItemText(v.List.GetCurrentItem())
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

	v.Parent.FocusedViewIndex = 4
	t.SetTitle(resourceName + " Details")
	t.AddInputField("Name", resourceName, 0, nil, nil).
		AddInputField("Resource Type", v.ResourceType, 0, nil, nil).
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
		logger.Println("failed to create resources client: ", err)
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
			logger.Println("failed to get the next page of", v.ResourceType, ":", err)
		}

		for _, resource := range page.Value {
			v.List.AddItem(*resource.Name, *resource.Location, 0, nil)
		}
	}

	return nil
}
