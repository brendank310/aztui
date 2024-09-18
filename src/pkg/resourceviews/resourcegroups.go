package resourceviews

import (
	"context"
	"fmt"

	"github.com/brendank310/aztui/pkg/layout"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var resourceGroupSelectItemFuncMap = map[string]func(*ResourceGroupListView) tview.Primitive{
  "SpawnResourceTypeListView": (*ResourceGroupListView).SpawnResourceTypeListView,
	"SpawnAKSClusterListView":     (*ResourceGroupListView).SpawnAKSClusterListView,
	"SpawnVirtualMachineListView": (*ResourceGroupListView).SpawnVirtualMachineListView,
}

type ResourceGroupListView struct {
	List           *tview.List
	StatusBarText  string
	ActionBarText  string
	SubscriptionID string
	Parent         *layout.AppLayout
}

func NewResourceGroupListView(appLayout *layout.AppLayout, subscriptionID string) *ResourceGroupListView {
	rg := ResourceGroupListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("Resource Groups (%v)", "F2")

	rg.List.SetBorder(true)
	rg.List.Box.SetTitle(title)
	rg.List.ShowSecondaryText(false)
	rg.ActionBarText = "## Select(Enter) ## | ## Exit(F12) ##"
	rg.SubscriptionID = subscriptionID
	rg.Parent = appLayout

	layout.InitKeyBindings[ResourceGroupListView, tview.List](appLayout, &rg, rg.List, resourceGroupSelectItemFuncMap, 1)

	return &rg
}

func (r *ResourceGroupListView) SpawnVirtualMachineListView() tview.Primitive {
	resourceGroup, _ := r.List.GetItemText(r.List.GetCurrentItem())
	vmList := NewVirtualMachineListView(r.Parent, r.SubscriptionID, resourceGroup)
	vmList.Update()
	return vmList.List
}

func (r *ResourceGroupListView) SpawnAKSClusterListView() tview.Primitive {
	resourceGroup, _ := r.List.GetItemText(r.List.GetCurrentItem())
	aksList := NewAKSClusterListView(r.Parent, r.SubscriptionID, resourceGroup)
	aksList.Update()

	return aksList.List

func (r *ResourceGroupListView) SpawnResourceTypeListView(resourceGroup string) tview.Primitive {
	rtList := NewResourceTypeListView(r.Parent, r.SubscriptionID, resourceGroup)
	rtList.Update()

	return rtList.List
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

	rgPager := rgClient.NewListPager(nil)
	for rgPager.More() {
		ctx := context.Background()
		page, err := rgPager.NextPage(ctx)
		if err != nil {
			return fmt.Errorf("failed to get next resource groups page: %v", err)
		}
		for _, rg := range page.Value {
			resourceGroup := *rg.Name
			r.List.AddItem(resourceGroup, "", 0, nil)
		}
	}

	return nil
}
