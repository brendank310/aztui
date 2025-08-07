package resourceviews

import (
	"context"
	"fmt"
	"log"

	"github.com/brendank310/aztui/pkg/cache"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
)

var aksClusterSelectItemFuncMap = map[string]func(*AKSClusterListView) tview.Primitive{
	"SpawnAKSClusterDetailView": (*AKSClusterListView).SpawnAKSClusterDetailView,
}

type AKSClusterListView struct {
	List           *tview.List
	StatusBarText  string
	ActionBarText  string
	SubscriptionID string
	ResourceGroup  string
	Parent         *AppLayout
}

func NewAKSClusterListView(appLayout *AppLayout, subscriptionID string, resourceGroup string) *AKSClusterListView {
	aks := AKSClusterListView{
		List: tview.NewList(),
	}

	appLayout.FocusedViewIndex = 2
	title := fmt.Sprintf("AKS Clusters (F%v)", appLayout.FocusedViewIndex+1)

	aks.List.SetBorder(true)
	aks.List.Box.SetTitle(title)
	aks.List.ShowSecondaryText(false)
	aks.ActionBarText = ""
	aks.SubscriptionID = subscriptionID
	aks.ResourceGroup = resourceGroup
	aks.Parent = appLayout

	aks.List.SetFocusFunc(func() {
		InitViewKeyBindings(&aks)
		aks.Update()
		aks.UpdateActionBar(aks.Parent.ActionBar)
	})

	return &aks
}

func (a *AKSClusterListView) UpdateActionBar(t *tview.TextView) {
	actionBarText := ""
	for _, view := range config.GConfig.Views {
		if view.Name == a.Name() {
			for _, action := range view.Actions {
				actionBarText += fmt.Sprintf("%v(%v) | ", action.Description, action.Key)
			}
			actionBarText = actionBarText[:len(actionBarText)-3] // Remove the last " | "
			break
		}
	}

	t.SetText(actionBarText)
}

func (v *AKSClusterListView) Name() string {
	return "AKSClusterListView"
}

func (v *AKSClusterListView) SetInputCapture(f func(event *tcell.EventKey) *tcell.EventKey) {
	v.List.SetInputCapture(f)
}

func (v *AKSClusterListView) CustomInputHandler() func(event *tcell.EventKey) *tcell.EventKey {
	return nil
}

func (v *AKSClusterListView) CallAction(action string) (tview.Primitive, error) {
	if actionFunc, ok := aksClusterSelectItemFuncMap[action]; ok {
		return actionFunc(v), nil
	}
	return nil, fmt.Errorf("no action for %s", action)
}

func (v *AKSClusterListView) AppendPrimitiveView(p tview.Primitive, takeFocus bool, width int) {
	v.Parent.AppendPrimitiveView(p, takeFocus, width)
}

func (v *AKSClusterListView) SpawnAKSClusterDetailView() tview.Primitive {
	aksClusterName, _ := v.List.GetItemText(v.List.GetCurrentItem())
	t := tview.NewForm()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatalf("Failed to obtain a credential: %v", err)
	}

	// Create a context
	ctx := context.Background()

	// Create a client to interact with AKS
	client, err := armcontainerservice.NewManagedClustersClient(v.SubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create AKS client: %v", err)
	}

	aksCluster, err := client.Get(ctx, v.ResourceGroup, aksClusterName, nil)
	if err != nil {
		log.Fatalf("Failed to get VM: %v", err)
	}

	t.SetTitle(aksClusterName + " Details")
	t.AddInputField("AKS Cluster Name", *aksCluster.Name, 0, nil, nil).
		AddInputField("Resource ID", *aksCluster.ID, 0, nil, nil).
		AddInputField("Location", *aksCluster.Location, 0, nil, nil)
	t.SetBorder(true)

	return t
}

func (v *AKSClusterListView) Update() error {
	// Use cache service for AKS cluster list
	cacheService := GetCacheService()
	if cacheService != nil {
		cacheKey := cache.GenerateAKSKey(v.SubscriptionID, v.ResourceGroup)
		
		// Try to get cached AKS clusters first
		data, err := cacheService.GetOrFetch(cacheKey, func() (interface{}, error) {
			return v.fetchAKSClusters()
		})
		
		if err != nil {
			return err
		}
		
		// Cast the cached data back to the expected type
		if clusters, ok := data.([]*armcontainerservice.ManagedCluster); ok {
			v.populateList(clusters)
			return nil
		}
	}
	
	// Fallback to direct fetch if cache service is not available
	clusters, err := v.fetchAKSClusters()
	if err != nil {
		return err
	}
	
	v.populateList(clusters)
	return nil
}

// fetchAKSClusters fetches AKS clusters from Azure API
func (v *AKSClusterListView) fetchAKSClusters() ([]*armcontainerservice.ManagedCluster, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to obtain a credential: %v", err)
	}

	// Create a context
	ctx := context.Background()

	// Create a client to interact with AKS
	client, err := armcontainerservice.NewManagedClustersClient(v.SubscriptionID, cred, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create AKS client: %v", err)
	}

	var clusters []*armcontainerservice.ManagedCluster

	// List AKS clusters in the specified resource group
	clusterListPager := client.NewListByResourceGroupPager(v.ResourceGroup, nil)

	// Iterate through the pager to fetch all AKS clusters
	for clusterListPager.More() {
		page, err := clusterListPager.NextPage(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to get the next page of AKS clusters: %v", err)
		}

		clusters = append(clusters, page.Value...)
	}

	return clusters, nil
}

// populateList populates the UI list with AKS cluster data
func (v *AKSClusterListView) populateList(clusters []*armcontainerservice.ManagedCluster) {
	v.List.Clear()

	if len(clusters) == 0 {
		v.List.AddItem("(No AKS clusters in resource group)", "", 0, nil)
		return
	}

	for _, cluster := range clusters {
		v.List.AddItem(*cluster.Name, *cluster.Properties.KubernetesVersion, 0, nil)
	}
}
