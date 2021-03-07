package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"os"
)

func main() {
	//
	// --- Step Outputs: Export Environment Variables for other Steps:
	// You can export Environment Variables for other Steps with
	//  envman, which is automatically installed by `bitrise setup`.
	// A very simple example:

	client := resty.New()

	resp, err := client.R().
		EnableTrace().
		Get("https://test-executor.smart-qa.io/")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp)

	//cmdLog, err := exec.Command("bitrise", "envman", "add", "--key", "EXAMPLE_STEP_OUTPUT", "--value", "the value you want to share").CombinedOutput()
	//if err != nil {
	//	fmt.Printf("Failed to expose output with envman, error: %#v | output: %s", err, cmdLog)
	//	os.Exit(1)
	//}

	// You can find more usage examples on envman's GitHub page
	//  at: https://github.com/bitrise-io/envman

	//
	// --- Exit codes:
	// The exit code of your Step is very important. If you return
	//  with a 0 exit code `bitrise` will register your Step as "successful".
	// Any non zero exit code will be registered as "failed" by `bitrise`.
	os.Exit(0)
}
