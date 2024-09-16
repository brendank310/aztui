package resourceviews

import (
	"context"
	"fmt"
	"log"

	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
)

type AKSListView struct {
	List *tview.List
	StatusBarText string
	ActionBarText string
	SubscriptionID string
	ResourceGroup string
}

func NewAKSListView(subscriptionID string, resourceGroup string) *AKSListView {
	aks := AKSListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("AKS Clusters (%v)", "F3")

	aks.List.SetBorder(true)
	aks.List.Box.SetTitle(title)
	aks.List.ShowSecondaryText(false)
	aks.ActionBarText = "## Subscription List(F1) ## | ## Resource Group List(F2) ## | ## Run Command(F5) ## | ## Exit(F12) ##"
	aks.SubscriptionID = subscriptionID
	aks.ResourceGroup = resourceGroup

	return &aks
}

func (v *AKSListView) Update(selectedFunc func()) error {
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
