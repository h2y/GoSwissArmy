package main

import (
	"fmt"
	"net/http"
)

func ErrorMsg(resp *http.Response, err error) string {
	ret := ""
	if resp != nil && resp.StatusCode != 200 {
		ret = fmt.Sprintf("code=%d", resp.StatusCode)
	}
	if err != nil {
		ret += fmt.Sprintf("err=%v", err)
	}
	return ret
}
