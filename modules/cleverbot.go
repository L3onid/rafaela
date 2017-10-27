package modules

import (
	"strings"
	"net/url"
)

import (
	"../commands"
	"net/http"
)

func Cleverbot(content string, userid string) (string) {
	v := url.Values{}
	v.Set("stimulus", comandos.GetArgs(content))
	v.Set("prevref", "")
	v.Set("lineRef", "")
	v.Set("icognoCheck", "rafaela-clever-bot-" + userid)
	en := v.Encode()
	req, err := http.NewRequest("POST", "http://app.cleverbot.com/webservicexml_ais_AYA", strings.NewReader(en))
	if err != nil {
		return ""
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", "Dalvik/2.1.0 (Linux; U; Android 7.1.2; MotoG3-TE Build/NJH47F)")
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	result, _ := url.QueryUnescape(resp.Header.Get("Cboutput"))
	return strings.ToLower(string(result[0])) + string(result[1:])
}
