package process

import (
	. "../jobQueue"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
)

func PushBear(text, desp string) {

	Logger.Info("收到发送pushbear的请求")
	if len(Config.Src.PushbearSendKey) > 0 {
		req, err := http.NewRequest("GET", "https://pushbear.ftqq.com/sub", nil)
		if err != nil {
			Logger.Warn("发送pushbear请求出错", zap.String("error", err.Error()))
		}
		q := req.URL.Query()
		q.Add("sendkey", Config.Src.PushbearSendKey)
		q.Add("text", text)
		q.Add("desp", desp)
		req.URL.RawQuery = q.Encode()
		var resp *http.Response
		//client := &http.Client{Timeout:time.Second * 1}
		resp, err = http.DefaultClient.Do(req)
		if err != nil {
			Logger.Warn("发送pushbear请求出错", zap.String("error", err.Error()))
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			// handle error
		}
		Logger.Info("发送pushbear请求结果", zap.String("res", string(body)))
		defer resp.Body.Close()
	}
}
