package main

import (
	"fmt"
	"os"
	"github.com/bitrise-tools/go-steputils/stepconf"
	"github.com/go-resty/resty/v2"
)

type Config struct {
	BuildPath          string          `env:"build_path,required"`
	APIKey             stepconf.Secret `env:"api_key,required"`
}

func main() {
	var cfg Config
	if err := stepconf.Parse(&cfg); err != nil {
		fmt.Println(err.Error())
	}

	client := resty.New()

	resp, err := client.R().
		EnableTrace().
		Get("https://test-executor.smart-qa.io/")

	if err != nil {
		fmt.Println(err)
	}

	fmt.Println(resp)
	os.Exit(0)
}
