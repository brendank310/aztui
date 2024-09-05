package main

import (
	_ "fmt"

	"strings"

	"github.com/brendank310/aztui/pkg/azcli"
	"github.com/brendank310/aztui/pkg/layout"
	"github.com/brendank310/aztui/pkg/resourceviews"

	"github.com/rivo/tview"
)

type AzTuiState struct {
	// Basic TUI variables
	*layout.AppLayout
}

func NewAzTuiState() *AzTuiState {
	// Base initialization
	a := AzTuiState{
		AppLayout: layout.NewAppLayout(),
	}

	subList := resourceviews.NewSubscriptionListView()
	a.AppLayout.AppendListView(subList.List)
	a.AppLayout.App.SetFocus(subList.List)
	subList.Update(func() {
		_, subscriptionID := subList.List.GetItemText(subList.List.GetCurrentItem())

		rgList := resourceviews.NewResourceGroupListView(subscriptionID)
		a.AppLayout.AppendListView(rgList.List)
		rgList.Update(func() {
			resourceGroupName, _ := rgList.List.GetItemText(rgList.List.GetCurrentItem())
			vmList := resourceviews.NewVirtualMachineListView(subscriptionID, resourceGroupName)
			a.AppLayout.AppendListView(vmList.List)
			vmList.Update(func() {
				cmdMap, err := azcli.GetResourceCommands("vm")
				if err != nil {
					panic(err)
				}

				cmdList := tview.NewList()
				cmdList.SetTitle("VM Commands")
				cmdList.SetBorder(true)
				for k, v := range cmdMap {
					cmdList.AddItem(k, v, 0, func() {
						cmdStr, _ := cmdList.GetItemText(cmdList.GetCurrentItem())
						vmName, _ := vmList.List.GetItemText(vmList.List.GetCurrentItem())
						args := []string{"vm", cmdStr, "-g", resourceGroupName, "-n", vmName}
						out, err := azcli.RunAzCommand(args, func(a []string, err error) error {
		if strings.HasPrefix(err.Error(), "ERROR: InvalidArgumentValue:") {
			newArgs := a
			cmdForm := tview.NewForm()

			missingArg := strings.Split(err.Error(), "field:")[1]
			cmdForm.AddInputField("Missing argument: " + missingArg, "", 0, nil, func(text string) {
				newArgs = append(newArgs, missingArg)
				newArgs = append(newArgs, text)
				_, _ = azcli.RunAzCommand(newArgs, nil)
			})
		}

		if strings.HasSuffix(err.Error(), "are required\n") {
			newArgs := a
			cmdForm := tview.NewForm()
			extractRequiredArgs := strings.Split(err.Error(), ":")[1]
			missingArgs := strings.Split(strings.TrimSuffix(strings.Replace(strings.Replace(extractRequiredArgs, "(", "", 1), ")", "", 1), " are required\n"), "|")

			for _, arg := range missingArgs {
				cmdForm.AddInputField("Missing argument: " + arg, "", 0, nil, func(text string) {
					newArgs = append(newArgs, arg)
					newArgs = append(newArgs, text)
				})
			}

			_, _ = azcli.RunAzCommand(newArgs, nil)
		}

		return nil
	})
						if err != nil {
							panic(err)
						}

						output := tview.NewTextView()
						output.SetTitle("Command Output")
						output.SetBorder(true)
						// Capture input here, maybe <Esc> to close cmd output?
						a.AppLayout.AppendTextView(output)
						output.Write([]byte(out))
					})
				}

				a.AppLayout.AppendListView(cmdList)
			})
		})
	})

	return &a
}


func main() {
	a := NewAzTuiState()

	if err := a.AppLayout.App.SetRoot(a.AppLayout.Grid, true).Run(); err != nil {
			panic(err)
	}
}
