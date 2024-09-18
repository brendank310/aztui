package resourceviews

import (
	"fmt"

	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/brendank310/aztui/pkg/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type ResourceTypeListView struct {
	List           *tview.List
	StatusBarText  string
	ActionBarText  string
	SubscriptionID string
	ResourceGroup  string
	Parent         *layout.AppLayout
	FuncMap        map[string]func(*ResourceTypeListView, string) tview.Primitive
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
	rt.FuncMap = make(map[string]func(*ResourceTypeListView, string) tview.Primitive)
	rt.FuncMap["SpawnAKSClusterListView"] = (*ResourceTypeListView).SpawnAKSClusterListView
	rt.FuncMap["SpawnVirtualMachineListView"] = (*ResourceTypeListView).SpawnVirtualMachineListView

	for _, action := range config.GConfig.Actions {
		if utils.GetTypeString[ResourceTypeListView]() == action.Type {
			// set the input capture
			rt.List.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
				Ch := rune(0)
				if event.Key() == tcell.KeyRune {
					Ch = event.Rune()
				}
				_ = Ch

				if method, exists := rt.FuncMap[action.Action]; exists {
					view := method(&rt, resourceGroup)
					if view != nil {
						rt.Parent.AppendPrimitiveView(view, action.TakeFocus, 1)
					}
					return nil
				}

				return event
			})
		}
	}

	return &rt
}

func (r *ResourceTypeListView) SpawnVirtualMachineListView(resourceType string) tview.Primitive {
	vmList := NewVirtualMachineListView(r.Parent, r.SubscriptionID, r.ResourceGroup)
	vmList.Update()

	return vmList.List
}

func (r *ResourceTypeListView) SpawnAKSClusterListView(resourceType string) tview.Primitive {
	aksList := NewAKSClusterListView(r.Parent, r.SubscriptionID, r.ResourceGroup)
	aksList.Update()

	return aksList.List
}
