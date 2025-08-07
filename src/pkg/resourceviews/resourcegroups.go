package resourceviews

import (
	"context"
	"fmt"
	"strings"

	"github.com/brendank310/aztui/pkg/cache"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var resourceGroupSelectItemFuncMap = map[string]func(*ResourceGroupListView) tview.Primitive{
	"SpawnResourceTypeListView":   (*ResourceGroupListView).SpawnResourceTypeListView,
	"SpawnVirtualMachineListView": (*ResourceGroupListView).SpawnVirtualMachineListView,
	"SpawnAKSClusterListView":     (*ResourceGroupListView).SpawnAKSClusterListView,
}

type ResourceGroupInfo struct {
	ResourceGroupName     string
	ResourceGroupLocation string
}

type ResourceGroupListView struct {
	List              *tview.List
	StatusBarText     string
	ActionBarText     string
	SubscriptionID    string
	Parent            *AppLayout
	ResourceGroupList *[]ResourceGroupInfo
}

func NewResourceGroupListView(appLayout *AppLayout, subscriptionID string) *ResourceGroupListView {
	rg := ResourceGroupListView{
		List: tview.NewList(),
	}
	appLayout.FocusedViewIndex = 1
	title := fmt.Sprintf("Resource Groups (F%v)", appLayout.FocusedViewIndex+1)

	rg.List.SetBorder(true)
	rg.List.Box.SetTitle(title)
	rg.List.ShowSecondaryText(true)
	rg.ActionBarText = ""
	rg.SubscriptionID = subscriptionID
	rg.Parent = appLayout

	rg.List.SetFocusFunc(func() {
		InitViewKeyBindings(&rg)
		rg.Update()
		rg.UpdateList(rg.Parent)
		rg.Parent.InputField.SetText("")
		rg.UpdateActionBar(rg.Parent.ActionBar)
	})

	return &rg
}

func (r *ResourceGroupListView) UpdateActionBar(t *tview.TextView) {
	actionBarText := ""
	for _, view := range config.GConfig.Views {
		if view.Name == r.Name() {
			for _, action := range view.Actions {
				actionBarText += fmt.Sprintf("%v(%v) | ", action.Description, action.Key)
			}
			actionBarText = actionBarText[:len(actionBarText)-3] // Remove the last " | "
			break
		}
	}

	t.SetText(actionBarText)
}

func (r *ResourceGroupListView) Name() string {
	return "ResourceGroupListView"
}

func (r *ResourceGroupListView) SetInputCapture(f func(event *tcell.EventKey) *tcell.EventKey) {
	r.List.SetInputCapture(f)
}

func (r *ResourceGroupListView) CustomInputHandler() func(event *tcell.EventKey) *tcell.EventKey {
	return nil
}

func (r *ResourceGroupListView) CallAction(action string) (tview.Primitive, error) {
	if actionFunc, ok := resourceGroupSelectItemFuncMap[action]; ok {
		return actionFunc(r), nil
	}
	return nil, fmt.Errorf("no action for %s", action)
}

func (r *ResourceGroupListView) AppendPrimitiveView(p tview.Primitive, takeFocus bool, width int) {
	r.Parent.AppendPrimitiveView(p, takeFocus, width)
}

func (r *ResourceGroupListView) SpawnResourceTypeListView() tview.Primitive {
	resourceGroup, _ := r.List.GetItemText(r.List.GetCurrentItem())
	// Remove previous views if exist starting from the one at index 2
	r.Parent.RemoveViews(2)
	rtList := NewResourceTypeListView(r.Parent, r.SubscriptionID, resourceGroup)
	return rtList.List
}

func (r *ResourceGroupListView) SpawnVirtualMachineListView() tview.Primitive {
	resourceGroup, _ := r.List.GetItemText(r.List.GetCurrentItem())
	// Remove previous views if exist starting from the one at index 2
	r.Parent.RemoveViews(2)
	vmList := NewVirtualMachineListView(r.Parent, r.SubscriptionID, resourceGroup)
	return vmList.List
}

func (r *ResourceGroupListView) SpawnAKSClusterListView() tview.Primitive {
	resourceGroup, _ := r.List.GetItemText(r.List.GetCurrentItem())
	// Remove previous views if exist starting from the one at index 2
	r.Parent.RemoveViews(2)
	aksList := NewAKSClusterListView(r.Parent, r.SubscriptionID, resourceGroup)
	return aksList.List
}

func (r *ResourceGroupListView) Update() error {
	// Use cache service for resource group list
	cacheService := GetCacheService()
	if cacheService != nil {
		cacheKey := cache.GenerateResourceGroupKey(r.SubscriptionID)
		
		// Try to get cached resource groups first
		data, err := cacheService.GetOrFetch(cacheKey, func() (interface{}, error) {
			return r.fetchResourceGroups()
		})
		
		if err != nil {
			return err
		}
		
		// Cast the cached data back to the expected type
		if resourceGroups, ok := data.([]ResourceGroupInfo); ok {
			r.ResourceGroupList = &resourceGroups
			r.populateList()
			return nil
		}
	}
	
	// Fallback to direct fetch if cache service is not available
	resourceGroups, err := r.fetchResourceGroups()
	if err != nil {
		return err
	}
	
	r.ResourceGroupList = &resourceGroups
	r.populateList()
	return nil
}

// fetchResourceGroups fetches resource groups from Azure API
func (r *ResourceGroupListView) fetchResourceGroups() ([]ResourceGroupInfo, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %v", err)
	}

	rgClient, err := armresources.NewResourceGroupsClient(r.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource groups client: %v", err)
	}

	var resourceGroups []ResourceGroupInfo

	rgPager := rgClient.NewListPager(nil)
	for rgPager.More() {
		ctx := context.Background()
		page, err := rgPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next resource groups page: %v", err)
		}
		for _, rg := range page.Value {
			resourceGroup := *rg.Name
			location := *rg.Location
			resourceGroups = append(resourceGroups, ResourceGroupInfo{resourceGroup, location})
		}
	}

	return resourceGroups, nil
}

// populateList populates the UI list with resource group data
func (r *ResourceGroupListView) populateList() {
	r.List.Clear()
	for _, resourceGroupInfo := range *r.ResourceGroupList {
		r.List.AddItem(resourceGroupInfo.ResourceGroupName, resourceGroupInfo.ResourceGroupLocation, 0, nil)
	}
}

func (r *ResourceGroupListView) UpdateList(layout *AppLayout) error {
	r.List.Clear()
	// Make filtering case insensitive
	filter := strings.ToLower(layout.InputField.GetText())
	for _, ResourceGroupInfo := range *r.ResourceGroupList {
		lowerCaseResourceGroupName := strings.ToLower(ResourceGroupInfo.ResourceGroupName)
		if strings.Contains(lowerCaseResourceGroupName, filter) {
			r.List.AddItem(ResourceGroupInfo.ResourceGroupName, ResourceGroupInfo.ResourceGroupLocation, 0, nil)
		}
	}
	return nil
}
