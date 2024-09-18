package main

import (
	_ "fmt"

	"os"
	_ "strings"

	_ "github.com/brendank310/aztui/pkg/azcli"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/brendank310/aztui/pkg/logger"
	"github.com/brendank310/aztui/pkg/resourceviews"

	"github.com/gdamore/tcell/v2"
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
		configPath = os.Getenv("HOME") + "/.config/aztui.yaml"
	}

	c, err := config.LoadConfig(configPath)
	if err != nil {
		panic(err)
	}

	a.Config = c

	subscriptionList := resourceviews.NewSubscriptionListView(a.AppLayout)
	if subscriptionList == nil {
		panic("unable to create a subscription list")
	}

	a.AppLayout.InputField.SetFinishedFunc(func(key tcell.Key) {
		subscriptionList.UpdateList(a.AppLayout)
		a.App.SetFocus(subscriptionList.List)
	})

	return &a
}

func main() {
	a := NewAzTuiState()

	if err := a.AppLayout.App.SetRoot(a.AppLayout.Grid, true).Run(); err != nil {
		panic(err)
	}
}
