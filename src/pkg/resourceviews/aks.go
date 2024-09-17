package resourceviews

import (
	"context"
	"fmt"
	"log"

	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"

	"github.com/rivo/tview"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerservice/armcontainerservice"
)

var aksSelectItemFuncMap = map[string]func(*AKSListView, string) tview.Primitive{
	"SpawnAKSDetailView": (*AKSListView).SpawnAKSDetailView,
}

type AKSListView struct {
	List           *tview.List
	StatusBarText  string
	ActionBarText  string
	SubscriptionID string
	ResourceGroup  string
	Parent	       *layout.AppLayout
}

func NewAKSListView(layout *layout.AppLayout, subscriptionID string, resourceGroup string) *AKSListView {
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
	aks.Parent = layout

	return &aks
}

func (a *AKSListView) SpawnAKSDetailView(clusterName string) tview.Primitive {
	t := tview.NewForm()
	t.SetTitle(clusterName + " Details")
	t.AddInputField("Cluster Name", clusterName, 0, nil, nil)
	t.SetBorder(true)

	return t
}

func callAKSMethodByName(view *AKSListView, methodName string, clusterName string) tview.Primitive {
	if method, exists := aksSelectItemFuncMap[methodName]; exists {
		return method(view, clusterName)
	} else {
		fmt.Printf("Method %s not found\n", methodName)
	}

	return nil
}

func (a *AKSListView) SelectItem(clusterName string) {
	symbolName := GetSymbolName()
	typeName := ExtractTypeName(symbolName)
	fnName := GetFunctionName(symbolName)

	for _, action := range config.GConfig.Actions {
		if typeName == action.Type && fnName == action.Condition {
			p := callAKSMethodByName(a, action.Action, clusterName)
			a.Parent.AppendPrimitiveView(p, action.TakeFocus, 3)
		}
	}
}

func (a *AKSListView) Update() error {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return fmt.Errorf("failed to obtain a credential: %v", err)
	}

	a.List.Clear()
	// Create a context
	ctx := context.Background()

	// Create a client to interact with AKS
	client, err := armcontainerservice.NewManagedClustersClient(a.SubscriptionID, cred, nil)
	if err != nil {
		log.Fatalf("failed to create AKS client: %v", err)
	}

	// List AKS clusters in the specified resource group
	clusterListPager := client.NewListByResourceGroupPager(a.ResourceGroup, nil)

	// Iterate through the pager to fetch all AKS clusters
	for clusterListPager.More() {
		page, err := clusterListPager.NextPage(ctx)
		if err != nil {
			log.Fatalf("failed to get the next page of AKS clusters: %v", err)
		}

		// Loop through the AKS clusters and print their details
		for _, cluster := range page.Value {
		        clusterName := *cluster.Name
			a.List.AddItem(clusterName,
					*cluster.Properties.KubernetesVersion,
					0,
					func() {
					a.SelectItem(clusterName)
			})
		}
	}

	return nil
}
