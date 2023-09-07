package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type ChanelStatus int8

const ChanelStatusPending ChanelStatus = 0
const ChanelStatusDone ChanelStatus = 1

type Chanel struct {
	URL string
	Msg string

	Status      ChanelStatus
	ErrMsg      string
	FetchTimeMS uint
	SegPath     []string
	SegTimeMS   uint
	trueURL     *url.URL

	DLBytes  uint
	DLTimeMS uint
	Score    uint
}

func GetChanels() []*Chanel {
	chanelSet := map[string]bool{}
	dumpCnt := 0
	chanels := make([]*Chanel, 0, 128)
	for _, masterList := range C.MasterLists {
		var masterListContent string
		{
			log.Println("start get masterList:", masterList)
			rsp, err := C.HttpClient.Get(masterList)
			if err != nil || rsp.StatusCode != 200 {
				log.Fatalf("fail to connect masterList %s: %s\n", masterList, ErrorMsg(rsp, err))
			}
			cont, err := io.ReadAll(rsp.Body)
			rsp.Body.Close()
			if err != nil {
				log.Fatalf("fail to get masterList %s: %v\n", masterList, err)
			}
			masterListContent = string(cont)
		}

		chanel := &Chanel{}
		chanels = append(chanels, chanel)
		for _, li := range strings.Split(masterListContent, "\n") {
			li = strings.TrimSpace(li)
			if li == "" {
				continue
			} else if strings.HasPrefix(li, "#") {
				if chanel.Msg == "" {
					chanel.Msg = li
				} else {
					chanel.Msg += "\n" + li
				}
			} else if _, ok := chanelSet[li]; ok {
				dumpCnt++
				chanel.Msg = ""
			} else {
				chanel.URL = li
				chanel = &Chanel{}
				chanels = append(chanels, chanel)
			}
		}
		chanels = chanels[:len(chanels)-1]
	}
	log.Printf("find %d chanels in %d MasterLists. %d dump chanels.\n", len(chanels), len(C.MasterLists), dumpCnt)
	return chanels
}

func (t *Chanel) getChanelContent() (ret string, err error) {
	rsp, err := C.HttpClient.Get(t.URL)
	if err != nil || rsp.StatusCode != 200 {
		err = fmt.Errorf(ErrorMsg(rsp, err))
		log.Println("fail to connect Chanel", t.URL, "due to", err)
		return
	}
	cont, err := io.ReadAll(rsp.Body)
	rsp.Body.Close()
	if err != nil {
		log.Println("fail to get Chanel", t.URL, "due to", err)
		return
	}
	ret = string(cont)

	t.trueURL = rsp.Request.URL
	if t.trueURL.String() != t.URL {
		log.Printf("DEBUG: redirect %s -> %s\n", t.URL, t.trueURL.String())
	}
	return
}

func (t *Chanel) GetSeq() error {
	startTime := time.Now().UnixMilli()
	chanelText, err := t.getChanelContent()
	t.FetchTimeMS = uint(time.Now().UnixMilli() - startTime)
	if err != nil {
		t.ErrMsg = fmt.Sprint(err)
		return err
	}

	// SegPath, SegTimeMS
	t.SegPath = []string{}
	for _, li := range strings.Split(chanelText, "\n") {
		li = strings.TrimSpace(li)
		if li == "" {
			continue
		} else if strings.HasPrefix(li, "#") {
			lis := strings.SplitN(li, "EXTINF:", 2)
			if len(lis) < 2 || t.SegTimeMS > 0 {
				continue
			}
			li = lis[1]
			lis = strings.SplitN(li, ",", 2)
			li = lis[0]
			lis = strings.SplitN(li, " ", 2)
			li = lis[0]
			segTime, _ := strconv.ParseFloat(li, 32)
			t.SegTimeMS = uint(1000 * segTime)
		} else {
			t.SegPath = append(t.SegPath, li)
		}
	}
	if t.SegTimeMS <= 0 {
		log.Println("WARN: no seg time found in", t.URL)
		t.SegTimeMS = 10 * 1000 // TODO: smart check seg Time
	}

	log.Printf("get %d segs:(%dms)\t%s\n", len(t.SegPath), t.FetchTimeMS, t.URL)
	return nil
}

func (t *Chanel) isSeqNotList(header http.Header) bool {
	ctype := header.Get("Content-Type")
	ctype = strings.ToLower(ctype)
	if ctype == "" {
		log.Println("WARN: can not detect seq or list by header")
		return true
	}
	//log.Println("DEBUG: Content-Type", ctype)

	checkMIME := []string{"application/vnd.apple.mpegurl", "text/plain", "application/x-mpegurl", "audio/x-mpegurl"}
	for i := range checkMIME {
		if strings.Contains(ctype, checkMIME[i]) {
			return false
		}
	}
	return true
}

func (t *Chanel) DownloadTest() (err error) {
	if t.ErrMsg != "" || t.SegTimeMS <= 0 {
		return nil
	}

	timeLimit := uint(float32(t.SegTimeMS)*(float32(C.Tolerate)/100+1) + 0.5)
	window := make([]byte, 1024)

	// DFS: test first stream
	for _, segPath := range t.SegPath {
		var seqUrl string
		{
			target, _ := url.Parse(segPath)
			seqUrl = t.trueURL.ResolveReference(target).String()
		}

		startTime := time.Now().UnixMilli()
		rsp, err := C.HttpClient.Get(seqUrl)
		if err != nil || rsp.StatusCode != 200 {
			t.ErrMsg = "fail to connect Steam:\t" + ErrorMsg(rsp, err)
			log.Println(t.ErrMsg, t.Msg)
			return err
		}

		if t.isSeqNotList(rsp.Header) {
			success := false
			for {
				length, err := rsp.Body.Read(window)
				t.DLBytes += uint(length)
				t.DLTimeMS = uint(time.Now().UnixMilli() - startTime)
				if err == io.EOF {
					success = true
					break
				} else if err != nil {
					t.ErrMsg = fmt.Sprint("fail to get Steam due to", err)
					log.Println("fail to get Steam", t.Msg, "due to", err)
					break
				} else if t.DLTimeMS > timeLimit {
					t.ErrMsg = fmt.Sprintf("slow Steam:\t%dKB/s", t.DLBytes*1000/1024/t.DLTimeMS)
					log.Println(t.ErrMsg, t.Msg)
					break
				}
			}
			rsp.Body.Close()

			// score
			if success {
				t.Score = uint(0.5 + float32(100*t.SegTimeMS)/float32(t.DLTimeMS))
				if int(t.Score) < 100-C.Tolerate {
					t.ErrMsg = fmt.Sprintf("low score %d%%", t.Score)
				}
			}
			break // only test one seq
		} else {
			rsp.Body.Close()
			log.Printf("checking subList:\t%s\n", segPath)
			subChanel := Chanel{
				URL: seqUrl,
				Msg: "#subChanel",
			}
			subChanel.GetSeq()
			subChanel.DownloadTest()

			if subChanel.ErrMsg == "" { // find available
				t.DLBytes, t.DLTimeMS, t.Score = subChanel.DLBytes, subChanel.DLTimeMS, subChanel.Score
				break
			} else {
				t.ErrMsg = subChanel.ErrMsg
			}
		}
	}

	if t.ErrMsg == "" {
		log.Printf("Available:\tstream speed score %d%%:\t%s\n", t.Score, t.Msg)
	}
	return
}
