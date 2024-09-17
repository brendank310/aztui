package main

import (
	_ "fmt"

	"os"

	_ "github.com/brendank310/aztui/pkg/azcli"
	"github.com/brendank310/aztui/pkg/config"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/brendank310/aztui/pkg/logger"
	"github.com/brendank310/aztui/pkg/resourceviews"

	_ "github.com/rivo/tview"
	"github.com/gdamore/tcell/v2"
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
		// configPath = os.Getenv("HOME") + "/.config/aztui.yaml"
		configPath = "/home/domi/aztui/.config/aztui.yaml"
		// /home/domi/aztui/.config/aztui.yaml
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
		a.AppLayout.Layout.
        // indexes := subscriptionList.List.FindItems(a.AppLayout.InputField.GetText(), "", false, true)
        // indexesLen := len(indexes)
        // logger.Println("in finished: ", indexes)
        // logger.Println("Item count: ", subscriptionList.List.GetItemCount())
        // runningItemCount := subscriptionList.List.GetItemCount()
		// itemCountInit := subscriptionList.List.GetItemCount()
        // for i := 0; i < indexesLen + 1; i++ {
		// 	k := i
        //     for j := i; j < itemCountInit; j++ {
        //         logger.Println("i: ", i, "j: ", j, " Item count: ", runningItemCount)
		// 		if (k >= runningItemCount) {
		// 			break
		// 		}
        //         mainText, _ := subscriptionList.List.GetItemText(k)
        //         logger.Println("i: ", i, "j: ", j, "current k (", k, "): ", mainText)
        //         if strings.Contains(strings.ToLower(mainText), a.AppLayout.InputField.GetText()) {
        //             logger.Println("BREAKING on k (", k, "): ", mainText)
        //             runningItemCount = subscriptionList.List.GetItemCount()
        //             break
        //         } else {
        //             logger.Println("i: ", i, "j: ", j, "REMOVED k (", k, "): ", mainText)
        //             subscriptionList.List.RemoveItem(k)
        //         }
        //         runningItemCount = subscriptionList.List.GetItemCount()
        //     }
        //     runningItemCount = subscriptionList.List.GetItemCount()
        //     logger.Println("Item count AFTER: ", runningItemCount)
        // }
        // indexes2 := subscriptionList.List.FindItems(a.AppLayout.InputField.GetText(), "", false, true)
        // logger.Println("end: ", indexes2)
        // a.App.SetFocus(subscriptionList.List)
    })

	return &a
}

func main() {
	a := NewAzTuiState()

	if err := a.AppLayout.App.SetRoot(a.AppLayout.Grid, true).Run(); err != nil {
		panic(err)
	}
}
