package process

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
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



func (f *CallbackJob) Do() error {

	statusCode, err := samplePost(f.SrcArray, f.CallbackUrl)
	if err != nil || statusCode != 200 {
		//fmt.Println("数据同步补偿数据回调地址调用不成功,重新push到失败队列", f)
		if err != nil {
			return err
		} else {
			return errors.New("回调返回statusCode:" + strconv.Itoa(statusCode))
		}
	} else {
		//fmt.Println("数据同步补偿数据回调成功", f)
	}
	//fmt.Println("数据同步补偿数据回调成功", f)

	return nil
}


func samplePost(srcArray map[string]interface{}, callbackUrl string) (int, error) {
	bytesData, err := json.Marshal(srcArray)
	if err != nil {
		fmt.Println(err.Error() )
		return -1, err
	}
	reader := bytes.NewReader(bytesData)
	request, err := http.NewRequest("POST", callbackUrl, reader)
	if err != nil {
		fmt.Println(err.Error())
		return -1, err
	}
	request.Header.Set("Content-Type", "application/json;charset=UTF-8")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return -1, err
	}
	defer resp.Body.Close()

	return resp.StatusCode, err
}


