package resourceviews

import (
	"context"
	"fmt"

	"github.com/brendank310/aztui/pkg/layout"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type SubscriptionListView struct {
	List             *tview.List
	StatusBarText    string
	ActionBarText    string
	Parent           *layout.AppLayout
	FuncMap          map[string]func(*SubscriptionListView) tview.Primitive
	InputHandlerList []func(event *tcell.EventKey) *tcell.EventKey
}

func NewSubscriptionListView(appLayout *layout.AppLayout) *SubscriptionListView {
	s := SubscriptionListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("Subscriptions (%v)", "F1")

	s.List.SetBorder(true)
	s.List.Box.SetTitle(title)
	s.ActionBarText = "## Select(Enter) ## | ## Exit(F12) ##"
	s.Parent = appLayout
	s.FuncMap = make(map[string]func(*SubscriptionListView) tview.Primitive)
	s.FuncMap["SpawnResourceGroupListView"] = (*SubscriptionListView).SpawnResourceGroupListView
	s.InputHandlerList = make([](func(event *tcell.EventKey) *tcell.EventKey), 0)
	InstallInputFunctions[SubscriptionListView](s)

	s.Update()

	return &s
}

func (s *SubscriptionListView) CallFunctionByName(name string) func(*SubscriptionListView) tview.Primitive {
	if method, exists := s.FuncMap[name]; exists {
		return method(s)
	}

	return nil
}

func (s *SubscriptionListView) AppendInputHandler(f func(event *tcell.EventKey) *tcell.EventKey) {
	s.InputHandlerList = append(s.InputHandlerList, f)
}

func (s *SubscriptionListView) SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey) {
	s.List.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		for _, fn := range s.InputHandlerList {
			if fn != nil {
				return fn(event)
			}
		}
		return event
	})
}

func (s *SubscriptionListView) AppendPrimitiveView(p tview.Primitive, takeFocus bool, width int) {
	s.AppendPrimitiveView(view, takeFocus, width)
}

func (s *SubscriptionListView) SpawnResourceGroupListView() tview.Primitive {
	_, subscriptionID := s.List.GetItemText(s.List.GetCurrentItem())
	rgList := NewResourceGroupListView(s.Parent, subscriptionID)
	rgList.Update()
	return rgList.List
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
			s.List.AddItem(subscriptionName, subscriptionID, 0, nil)
		}
	}

	return nil
}
