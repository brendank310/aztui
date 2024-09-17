package main

import (
	_ "fmt"

	_ "strings"
	"os"

	_ "github.com/brendank310/aztui/pkg/azcli"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/brendank310/aztui/pkg/logger"
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
	err := logger.InitLogger()
	if err != nil {
		panic(err)
	}
	a := AzTuiState{
		AppLayout: layout.NewAppLayout(),
	}

	configPath := os.Getenv("AZTUI_CONFIG_PATH")
	if configPath == "" {
		configPath = "~/.config/aztui.yaml"
	}

	c, err := config.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}

	a.Config = c

	subList := resourceviews.NewSubscriptionListView(a.AppLayout)
	a.AppLayout.AppendPrimitiveView(subList.List)
	subList.Update(func() {
		_, subscriptionID := subList.List.GetItemText(subList.List.GetCurrentItem())
		//a.AppLayout.AppendPrimitiveView()
		subList.SelectItem(subscriptionID)
	})

	return &a
}

func main() {
	a := NewAzTuiState()

	if err := a.AppLayout.App.SetRoot(a.AppLayout.Grid, true).Run(); err != nil {
		panic(err)
	}
}
