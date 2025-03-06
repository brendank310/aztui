package resourceviews

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/brendank310/aztui/pkg/azcli"
)

type VMCommandListView struct {
	List          *tview.List
	StatusBarText string
	ActionBarText string
	ResourceGroup string
	VM            string
}

func NewVMCommandListView(resourceGroupName string, VM string) *VMCommandListView {
	s := VMCommandListView{
		List: tview.NewList(),
	}

	title := fmt.Sprintf("VM Commands (%v)", "F4")

	s.List.SetBorder(true)
	s.List.Box.SetTitle(title)
	s.ActionBarText = "## Select(Enter) ## | ## Exit(F12) ##"
	s.VM = VM
	s.ResourceGroup = resourceGroupName

	s.Update()

	return &s
}

// func (a *VMCommandListView) GetActionBarText() string {
// 	actionBarText := ""
// 	for _, view := range config.GConfig.Views {
// 		if view.Name == a.Name() {
// 			for _, action := range view.Actions {
// 				actionBarText += fmt.Sprintf("%v(%v) | ", action.Action, action.Key)
// 			}
// 			actionBarText = actionBarText[:len(actionBarText)-3] // Remove the last " | "
// 			break
// 		}
// 	}

// 	return actionBarText
// }

func (s *VMCommandListView) Update() error {
	cmdMap, err := azcli.GetResourceCommands("vm")
	if err != nil {
		panic(err)
	}

	for k, v := range cmdMap {
		s.List.AddItem(k, v, 0, func() {
			cmdStr, _ := s.List.GetItemText(s.List.GetCurrentItem())
			vmName := s.VM
			args := []string{"vm", cmdStr, "-g", s.ResourceGroup, "-n", vmName}
			out, err := azcli.RunAzCommand(args, func(a []string, err error) error {
				if strings.HasPrefix(err.Error(), "ERROR: InvalidArgumentValue:") {
					newArgs := a
					cmdForm := tview.NewForm()

					missingArg := strings.Split(err.Error(), "field:")[1]
					cmdForm.AddInputField("Missing argument: "+missingArg, "", 0, nil, func(text string) {
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
						cmdForm.AddInputField("Missing argument: "+arg, "", 0, nil, func(text string) {
							newArgs = append(newArgs, arg)
							newArgs = append(newArgs, text)
						})
					}

					_, _ = azcli.RunAzCommand(newArgs, nil)
				}

				return nil
			})
			_ = out
			_ = err
		})
	}

	return nil
}
