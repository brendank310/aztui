package resourceviews

import (
	"context"
	"fmt"
	"log"

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
