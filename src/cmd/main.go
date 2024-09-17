package main

import (
	_ "fmt"

	_ "strings"

	_ "github.com/brendank310/aztui/pkg/azcli"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/brendank310/aztui/pkg/resourceviews"

	_ "github.com/rivo/tview"
)

type AzTuiState struct {
	// Basic TUI variables
	*layout.AppLayout
	config.Config
}

func NewAzTuiState() *AzTuiState {
	// Base initialization
	a := AzTuiState{
		AppLayout: layout.NewAppLayout(),
	}

	c, err := config.LoadConfig("/etc/aztui.yaml")
	if err != nil {
		panic(err)
	}

	a.Config = c

	subList := resourceviews.NewSubscriptionListView(a.AppLayout)
	subList.Update()
	a.AppLayout.AppendPrimitiveView(subList.List)

	return &a
}


func main() {
	a := NewAzTuiState()

	if err := a.AppLayout.App.SetRoot(a.AppLayout.Grid, true).Run(); err != nil {
			panic(err)
	}
}
