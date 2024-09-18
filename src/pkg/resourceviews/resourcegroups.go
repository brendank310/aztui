package resourceviews

import (
	"context"
	"fmt"

	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/brendank310/aztui/pkg/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

type ResourceGroupListView struct {
	List           *tview.List
	StatusBarText  string
	ActionBarText  string
	SubscriptionID string
	Parent         *layout.AppLayout
	FuncMap        map[string]func(*ResourceGroupListView) tview.Primitive
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
	rg.FuncMap = make(map[string]func(*ResourceGroupListView) tview.Primitive)
	rg.FuncMap["SpawnResourceTypeListView"] = (*ResourceGroupListView).SpawnResourceTypeListView
	rg.FuncMap["SpawnAKSClusterListView"] = (*ResourceGroupListView).SpawnAKSClusterListView
	rg.FuncMap["SpawnVirtualMachineListView"] = (*ResourceGroupListView).SpawnVirtualMachineListView
	rg.Update()
	for _, action := range config.GConfig.Actions {
		targetType := utils.GetTypeString[ResourceGroupListView]()
		if action.Type == targetType {
			t := action.Type
			a := action.Action
			k := action.Key.Key
			rg.List.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				if t == targetType && k == event.Key() {
					if method, exists := rg.FuncMap[a]; exists {
						view := method(&rg)
						if view != nil {
							rg.Parent.AppendPrimitiveView(view, action.TakeFocus, action.Width)
						}
						return nil
					}
					return event
				}
				return event
			})
		}
	}

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
}

func (r *ResourceGroupListView) SpawnResourceTypeListView() tview.Primitive {
	resourceGroup, _ := r.List.GetItemText(r.List.GetCurrentItem())
	rtList := NewResourceTypeListView(r.Parent, r.SubscriptionID, resourceGroup)

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
