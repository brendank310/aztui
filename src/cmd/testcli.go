package main

import (
	"log"
	"os"
	"bufio"
	"strings"

	"github.com/brendank310/aztui/pkg/azcli"
)

func main() {
	subcommand := "show"
	subscriptionID := "16df4391-b8c8-4ff8-bd7d-6c725bfab3f6"
	resourceGroup := "brendank310-orgmode-rg"
	//vmName := "testvm"
	args := []string{"vm",
		subcommand,
	 	"--subscription",
	 	strings.TrimSpace(subscriptionID),
		"--resource-group",
		resourceGroup,
	 	//"--name",
	 	//strings.TrimSpace(vmName),
	}
	rc, err := azcli.GetResourceCommands(args[0])
	if err != nil {
		log.Fatalf("%v\n%v", rc, err)
	}

	_, err = azcli.RunAzCommand(args, func(a []string, err error) error {
		if strings.HasPrefix(err.Error(), "ERROR: InvalidArgumentValue:") {
			newArgs := a
			reader := bufio.NewReader(os.Stdin) // Create a new reader
			missingArg := strings.Split(err.Error(), "field:")[1]
			log.Printf("Missing argument %v - Enter now: ", missingArg)
			rg, _ := reader.ReadString('\n') // Read input until newline
			newArgs = append(newArgs, missingArg)
			newArgs = append(newArgs, rg)

			_, e := azcli.RunAzCommand(newArgs, nil)
			if e != nil {
				return e
			}
		}

		if strings.HasSuffix(err.Error(), "are required\n") {
			newArgs := a
			reader := bufio.NewReader(os.Stdin) // Create a new reader

			extractRequiredArgs := strings.Split(err.Error(), ":")[1]
			missingArgs := strings.Split(strings.TrimSuffix(strings.Replace(strings.Replace(extractRequiredArgs, "(", "", 1), ")", "", 1), " are required\n"), "|")

			for _, arg := range missingArgs {
				log.Printf("Enter value for %v: ", arg)
				value, _ := reader.ReadString('\n') // Read input until newlin

				newArgs = append(newArgs, strings.TrimSpace(arg))
				newArgs = append(newArgs, strings.TrimSpace(value))

				break
			}

			stdout, e := azcli.RunAzCommand(newArgs, nil)
			if e != nil {
				return e
			}

			log.Println(stdout)
		}

		return nil
	})
}
