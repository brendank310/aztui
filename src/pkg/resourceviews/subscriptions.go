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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armsubscriptions"
)

var subscriptionSelectItemFuncMap = map[string]func(*SubscriptionListView) tview.Primitive{
	"SpawnResourceGroupListView": (*SubscriptionListView).SpawnResourceGroupListView,
}

type SubscriptionInfo struct {
	SubscriptionName string
	SubscriptionID   string
}

type SubscriptionListView struct {
	List                  *tview.List
	StatusBarText         string
	ActionBarText         string
	Parent                *AppLayout
	SubscriptionList      *[]SubscriptionInfo
	ResourceGroupListView *ResourceGroupListView
}

func NewSubscriptionListView(appLayout *AppLayout) *SubscriptionListView {
	s := SubscriptionListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("Subscriptions (%v)", "F1")

	s.List.SetBorder(true)
	s.List.Box.SetTitle(title)
	s.ActionBarText = ""
	s.Parent = appLayout

	s.List.SetFocusFunc(func() {
		InitViewKeyBindings(&s)
		s.Update()
		s.UpdateList(s.Parent)
		s.Parent.InputField.SetText("")
		s.UpdateActionBar(s.Parent.ActionBar)
	})

	appLayout.AppendPrimitiveView(s.List, true, 1)
	return &s
}

func (s *SubscriptionListView) UpdateActionBar(t *tview.TextView) {
	actionBarText := ""
	for _, view := range config.GConfig.Views {
		if view.Name == s.Name() {
			for _, action := range view.Actions {
				actionBarText += fmt.Sprintf("%v(%v) | ", action.Description, action.Key)
			}
			actionBarText = actionBarText[:len(actionBarText)-3] // Remove the last " | "
			break
		}
	}

	t.SetText(actionBarText)
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
	s.Parent.RemoveViews(1)
	rgList := NewResourceGroupListView(s.Parent, subscriptionID)
	s.ResourceGroupListView = rgList
	rgList.UpdateActionBar(rgList.Parent.ActionBar)
	return rgList.List
}

func (s *SubscriptionListView) Update() error {
	// Use cache service for subscription list
	cacheService := GetCacheService()
	if cacheService != nil {
		cacheKey := cache.GenerateSubscriptionKey()
		
		// Try to get cached subscriptions first
		data, err := cacheService.GetOrFetch(cacheKey, func() (interface{}, error) {
			return s.fetchSubscriptions()
		})
		
		if err != nil {
			return err
		}
		
		// Cast the cached data back to the expected type
		if subscriptions, ok := data.([]SubscriptionInfo); ok {
			s.SubscriptionList = &subscriptions
			s.populateList()
			return nil
		}
	}
	
	// Fallback to direct fetch if cache service is not available
	subscriptions, err := s.fetchSubscriptions()
	if err != nil {
		return err
	}
	
	s.SubscriptionList = &subscriptions
	s.populateList()
	return nil
}

// fetchSubscriptions fetches subscriptions from Azure API
func (s *SubscriptionListView) fetchSubscriptions() ([]SubscriptionInfo, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %v", err)
	}

	subClient, err := armsubscriptions.NewClient(cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriptions client: %v", err)
	}

	var subscriptions []SubscriptionInfo

	// List subscriptions
	subPager := subClient.NewListPager(nil)
	ctx := context.Background()
	for subPager.More() {
		page, err := subPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get next subscriptions page: %v", err)
		}
		for _, subscription := range page.Value {
			subscriptionID := *subscription.SubscriptionID
			subscriptionName := *subscription.DisplayName
			subscriptions = append(subscriptions, SubscriptionInfo{subscriptionName, subscriptionID})
		}
	}

	return subscriptions, nil
}

// populateList populates the UI list with subscription data
func (s *SubscriptionListView) populateList() {
	s.List.Clear()
	for _, subscriptionInfo := range *s.SubscriptionList {
		s.List.AddItem(subscriptionInfo.SubscriptionName, subscriptionInfo.SubscriptionID, 0, nil)
	}
}

func (s *SubscriptionListView) UpdateList(layout *AppLayout) error {
	s.List.Clear()
	// Make filtering case insensitive
	filter := strings.ToLower(layout.InputField.GetText())
	for _, SubscriptionInfo := range *s.SubscriptionList {
		lowerCaseSubscriptionName := strings.ToLower(SubscriptionInfo.SubscriptionName)
		if strings.Contains(lowerCaseSubscriptionName, filter) {
			s.List.AddItem(SubscriptionInfo.SubscriptionName, SubscriptionInfo.SubscriptionID, 0, nil)
		}
	}
	return nil
}
