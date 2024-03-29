package main

import (
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/bitrise-io/go-utils/log"
	"github.com/bitrise-tools/go-steputils/stepconf"
	"github.com/go-resty/resty/v2"
	"github.com/mholt/archiver/v3"
	"io/ioutil"
	"os"
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

// zip 引数のパス配下のファイルを zip で圧縮する。
// 同名の zip が存在する場合は削除してから圧縮する。
func zip(path, zipName string) error {
	if err := os.RemoveAll(zipName); err != nil {
		msg := fmt.Sprintf("cannot remove %s. err: %s", zipName, err.Error())
		return errors.New(msg)
	}

	if err := archiver.Archive([]string{path}, zipName); err != nil {
		msg := fmt.Sprintf("cannot archive %s, err: %s", zipName, err.Error())
		return errors.New(msg)
	}

	return nil
}

type RunWithAppBody struct {
	AppURL string `json:"appURL"`
	ApiKey string `json:"apiKey"`
}

type RunWithAppReq struct {
	Data RunWithAppBody `json:"data"`
}

type reqHeader = map[string]string

type RunWithAppErrorRes struct {
	Error RunWithAppError
}

type RunWithAppError struct {
	Message string
	Status  string
}

func main() {
	var cfg Config
	bucketName := "smart-qa-prod.appspot.com"
	tmpArchiveFileName := "archive.app.zip"
	remotePath := fmt.Sprintf("tmp/builds/%s.app.zip", hash())
	uploadURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o?name=%s", bucketName, remotePath)
	appURL := fmt.Sprintf("https://firebasestorage.googleapis.com/v0/b/%s/o/%s", bucketName, remotePath)
	functionsURL := "https://asia-northeast1-smart-qa-prod.cloudfunctions.net/runWithApp"

	if err := stepconf.Parse(&cfg); err != nil {
		failed(err.Error())
	}

	if err := zip(cfg.BuildPath, tmpArchiveFileName); err != nil {
		failed(err.Error())
	}

	fileBytes, err := ioutil.ReadFile(tmpArchiveFileName)
	if err != nil {
		msg := fmt.Sprintf("err: %s, path: %s", err.Error(), tmpArchiveFileName)
		failed(msg)
	}

	client := resty.New()

	uploadRes, err := client.R().SetBody(fileBytes).SetHeaders(reqHeader{"Content-Type": "application/zip"}).Post(uploadURL)
	if err != nil {
		msg := fmt.Sprintf("failed to upload zip. err: %s", err.Error())
		failed(msg)
	}

	if uploadRes.StatusCode() > 299 {
		// TODO: 詳細なエラーを表示すること。（HTTP status に応じたエラーの出し分けとか。）
		msg := "failed to upload zip."
		failed(msg)
	}

	log.Printf(fmt.Sprintf("upload complete! file: %s", appURL))

	body := RunWithAppReq{
		Data: RunWithAppBody{
			AppURL: appURL,
			ApiKey: string(cfg.APIKey),
		},
	}

	j, err := json.Marshal(body)
	if err != nil {
		failed(err.Error())
	}

	res, err := client.R().SetBody(string(j)).SetHeaders(reqHeader{"Content-Type": "application/json"}).SetError(&RunWithAppErrorRes{}).Post(functionsURL)

	if err != nil {
		failed(err.Error())
	}

	// Status Code が 300 以上の場合は異常とみなす。go-resty の制約上、`RunWithAppErrorRes` というところにエラーが入ってくることに注意。
	if res.StatusCode() > 299 {
		errMessage := res.Error().(*RunWithAppErrorRes).Error.Message
		failed(errMessage)
	}

	log.Printf("successfully start to run test!")

	os.Exit(0)
}
