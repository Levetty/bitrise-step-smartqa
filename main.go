package main

import (
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"os"
	"os/exec"
	"time"
)

type Config struct {
	BuildPath string          `env:"build_path,required"`
	APIKey    stepconf.Secret `env:"api_key,required"`
}

// hash 現在時刻をシードとしたハッシュを作成。
func hash() string {
	now := time.Now().String()
	sha1 := sha1.Sum([]byte(now))
	return fmt.Sprintf("%x", sha1)
}

func failed(cause string) {
	log.Errorf(cause)
	os.Exit(1)
}

type UploadBuildBody struct {
	AppURL string `json:"appURL"`
	ApiKey string `json:"apiKey"`
}

type UploadBuildReq struct {
	Data UploadBuildBody `json:"data"`
}

type reqHeader = map[string]string

func main() {
	var cfg Config
	bucketName := "e2e-test-dev.appspot.com"
	tmpArchiveFileName := "archive.app.zip"
	remotePath := fmt.Sprintf("tmp/builds/%s.app.zip", hash())
	uploadURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o?name=%s", bucketName, remotePath)
	appURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s", bucketName, remotePath)
	functionsURL := "https://asia-northeast1-e2e-test-dev.cloudfunctions.net/runWithApp"

	if err := stepconf.Parse(&cfg); err != nil {
		failed(err.Error())
	}

	fmt.Println(exec.Command("zip", "-r", tmpArchiveFileName, cfg.BuildPath).String())
	if _ ,err := exec.Command("zip", tmpArchiveFileName, cfg.BuildPath).Output(); err != nil {
		failed(err.Error())
	}

	fileBytes, err := ioutil.ReadFile(tmpArchiveFileName)
	if err != nil {
		msg := fmt.Sprintf("err: %s, path: %s", err.Error(), cfg.BuildPath)
		failed(msg)
	}

	client := resty.New()

	res, err := client.R().SetBody(fileBytes).SetHeaders(reqHeader{"Content-Type": "application/zip"}).Post(uploadURL)
	if err != nil {
		failed(err.Error())
	}

	log.Printf(res.String())

	body := UploadBuildReq{
		Data: UploadBuildBody{
			AppURL: appURL,
			ApiKey: string(cfg.APIKey),
		},
	}

	j, err := json.Marshal(body)
	if err != nil {
		failed(err.Error())
	}

	resp, err := client.R().SetBody(string(j)).SetHeaders(reqHeader{"Content-Type": "application/json"}).Post(functionsURL)
	if err != nil {
		failed(err.Error())
	}

	log.Printf(resp.String())
	os.Exit(0)
}
