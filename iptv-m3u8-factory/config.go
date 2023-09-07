package main

import (
	"flag"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"
)

type Config struct {
	MasterLists []string
	Tolerate    int
	Output      string
	Coroutine   int

	HttpClient *http.Client
}

var C Config

func Init() {
	lists := flag.String("lists", "https://iptv-org.github.io/iptv/languages/zho.m3u", "m3u8 lists split by |")
	flag.IntVar(&C.Tolerate, "tolerate", 20, "allowed block percents of the stream")
	flag.IntVar(&C.Coroutine, "coroutine", 4, "coroutine to test lists")
	flag.StringVar(&C.Output, "output", "./iptv.m3u8", "output file path")
	flag.Parse()

	// MasterLists
	for _, list := range strings.Split(*lists, "|") {
		C.MasterLists = append(C.MasterLists, strings.TrimSpace(list))
	}

	// HttpClient
	jar, _ := cookiejar.New(nil)
	C.HttpClient = &http.Client{
		Jar:     jar,
		Timeout: 60 * time.Second,
	}
}
