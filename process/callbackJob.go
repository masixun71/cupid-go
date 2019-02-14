package process

import (
	. "../jobQueue"
	"bytes"
	"encoding/json"
	"errors"
	"go.uber.org/zap"
	"net/http"
	"strconv"
	"time"
)

type CallbackJob struct {
	Types       uint8
	SrcArray    map[string]interface{}
	CallbackUrl string
	RetireTime  int
}

func (f *CallbackJob) GetRetireTime() int {
	return f.RetireTime
}

func (f *CallbackJob) IncrRetireTime() {
	f.RetireTime++
}

func (f *CallbackJob) String() string {

	bytes,_ := json.Marshal(f)

	return string(bytes)
}


func (f *CallbackJob) Do() error {

	statusCode, err := samplePost(f.SrcArray, f.CallbackUrl)
	if err != nil || statusCode != 200 {
		Logger.Error("数据同步补偿数据回调地址调用不成功,重新push到失败队列", zap.String("job", f.String()), zap.String("err", err.Error()))
		if err != nil {
			return err
		} else {
			return errors.New("回调返回statusCode:" + strconv.Itoa(statusCode))
		}
	} else {
		Logger.Info("数据同步补偿数据回调地址调用成功", zap.String("job", f.String()))
	}
	Logger.Info("数据同步补偿数据回调地址调用成功", zap.String("job", f.String()))

	return nil
}


func samplePost(srcArray map[string]interface{}, callbackUrl string) (int, error) {
	bytesData, err := json.Marshal(srcArray)
	if err != nil {
		Logger.Warn("解析回调发送的srcArray出错", zap.String("error", err.Error()))
		return -1, err
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", callbackUrl, reader)
	if err != nil {
		Logger.Warn("发送回调请求出错", zap.String("error", err.Error()))
		return -1, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := &http.Client{Timeout:time.Second * 3}
	resp, err := client.Do(request)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, err
}


