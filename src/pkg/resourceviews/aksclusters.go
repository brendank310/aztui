package resourceviews

import (
	"context"
	"fmt"
	"log"

	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/brendank310/aztui/pkg/utils"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
)

type AKSClusterListView struct {
	List           *tview.List
	StatusBarText  string
	ActionBarText  string
	SubscriptionID string
	ResourceGroup  string
	Parent         *layout.AppLayout
	FuncMap        map[string]func(*AKSClusterListView) tview.Primitive
}

func NewAKSClusterListView(appLayout *layout.AppLayout, subscriptionID string, resourceGroup string) *AKSClusterListView {
	aks := AKSClusterListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("AKS Clusters (%v)", "F4")

	aks.List.SetBorder(true)
	aks.List.Box.SetTitle(title)
	aks.List.ShowSecondaryText(false)
	aks.ActionBarText = "## Subscription List(F1) ## | ## Resource Group List(F2) ## | ## Run Command(F5) ## | ## Exit(F12) ##"
	aks.SubscriptionID = subscriptionID
	aks.ResourceGroup = resourceGroup
	aks.Parent = appLayout
	aks.FuncMap["SpawnAKSClusterDetailView"] = (*AKSClusterListView).SpawnAKSClusterDetailView
	for _, action := range config.GConfig.Actions {
		if utils.GetTypeString[AKSClusterListView]() == action.Type {
			// set the input capture
			aks.List.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {

				Ch := rune(0)
				if event.Key() == tcell.KeyRune {
					Ch = event.Rune()
				}
				_ = Ch

				if method, exists := aks.FuncMap[action.Action]; exists {
					view := method(&aks)
					if view != nil {
						aks.Parent.AppendPrimitiveView(view, action.TakeFocus, 3)
					}
					return nil
				}
				return event
			})
		}
	}

	return &aks
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
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}

	v.List.Clear()
	// Create a context
	ctx := context.Background()

	// Create a client to interact with AKS
	client, err := armcontainerservice.NewManagedClustersClient(v.SubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create AKS client: %v", err)
	}

	// List AKS clusters in the specified resource group
	clusterListPager := client.NewListByResourceGroupPager(v.ResourceGroup, nil)

	// Check if the pager is empty
	if !clusterListPager.More() {
		v.List.AddItem("(No AKS clusters in resource group)", "", 0, nil)
	}

	// Iterate through the pager to fetch all AKS clusters
	for clusterListPager.More() {
		page, err := clusterListPager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to get the next page of AKS clusters: %v", err)
		}

		// Loop through the AKS clusters and print their details
		for _, cluster := range page.Value {
			v.List.AddItem(*cluster.Name, *cluster.Properties.KubernetesVersion, 0, nil)
		}
	}

	return nil
}
