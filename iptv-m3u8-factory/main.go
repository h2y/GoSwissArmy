package main

import (
	"fmt"
	"log"
	"os"
	"sync/atomic"
)

func main() {
	Init()

	chanels := GetChanels()

	var chanelsI int32 = -1
	for i := 0; i < C.Coroutine; i++ {
		go func() {
			for {
				ci := atomic.AddInt32(&chanelsI, 1)
				if int(ci) >= len(chanels) {
					return
				}
				chanel := chanels[ci]

				chanel.GetSeq()
				chanel.DownloadTest()
				chanel.Status = ChanelStatusDone
			}
		}()
	}

	// output
	out, err := os.Create(C.Output)
	if err != nil {
		log.Fatal("Create output file error:", err)
		return
	}
	defer out.Close()
	out.WriteString("#EXTM3U\n")

	availableCnt := 0
	for i := 0; i < len(chanels); {
		chanel := chanels[i]
		switch chanel.Status {
		case ChanelStatusPending:
		case ChanelStatusDone:
			if chanel.ErrMsg == "" {
				out.WriteString(fmt.Sprintf("# Score: %d%%\n", chanel.Score))
				out.WriteString(chanel.Msg + "\n" + chanel.URL + "\n")
				availableCnt++
			} else {
				out.WriteString(fmt.Sprintf("# Bad: %s\n", chanel.ErrMsg))
				out.WriteString(chanel.Msg + "\n#" + chanel.URL + "\n")
			}
			i++
		}
	}

	log.Printf("Finished. %d/%d chanels available.", availableCnt, len(chanels))
}
