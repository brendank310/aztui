package resourceviews

import (
	"context"
	"fmt"

	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

type SubscriptionListView struct {
	List *tview.List
	StatusBarText string
	ActionBarText string
}

func NewSubscriptionListView() *SubscriptionListView {
	s := SubscriptionListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("Subscriptions (%v)", "F1")

	s.List.SetBorder(true)
	s.List.Box.SetTitle(title)
	s.ActionBarText = "## Select(Enter) ## | ## Exit(F12) ##"

	return &s
}

func ProcessSelection(subscriptionID string) error {
	return nil
}

func (s *SubscriptionListView) Update(selectedFunc func()) error {
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
			subID := *subscription.SubscriptionID
			subName := *subscription.DisplayName
			s.List.AddItem(subName, subID, 0, selectedFunc)
		}
	}

	return nil
}
