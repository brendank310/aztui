package resourceviews

import (
	"context"
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

var subscriptionSelectItemFuncMap = map[string]func(*SubscriptionListView) tview.Primitive{
	"SpawnResourceGroupListView": (*SubscriptionListView).SpawnResourceGroupListView,
}

type SubscriptionListView struct {
	List          *tview.List
	StatusBarText string
	ActionBarText string
	Parent        *AppLayout
}

func NewSubscriptionListView(appLayout *AppLayout) *SubscriptionListView {
	s := SubscriptionListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("Subscriptions (%v)", "F1")

	s.List.SetBorder(true)
	s.List.Box.SetTitle(title)
	s.ActionBarText = "## Select(Enter) ## | ## Exit(F12) ##"
	s.Parent = appLayout

	InitViewKeyBindings(&s)

	s.Update()
	appLayout.AppendPrimitiveView(s.List, true, 1)
	return &s
}

func (s *SubscriptionListView) Name() string {
	return "SubscriptionListView"
}

func (s *SubscriptionListView) SetInputCapture(f func(event *tcell.EventKey) *tcell.EventKey) {
	s.List.SetInputCapture(f)
}

func (s *SubscriptionListView) CustomInputHandler() func(event *tcell.EventKey) *tcell.EventKey {
	return nil
}

func (s *SubscriptionListView) CallAction(action string) (tview.Primitive, error) {
	if actionFunc, ok := subscriptionSelectItemFuncMap[action]; ok {
		return actionFunc(s), nil
	}
	return nil, fmt.Errorf("no action for %s", action)
}

func (s *SubscriptionListView) AppendPrimitiveView(p tview.Primitive, takeFocus bool, width int) {
	s.Parent.AppendPrimitiveView(p, takeFocus, width)
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
