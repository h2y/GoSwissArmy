package main

import (
	"flag"
	"fmt"
	"github.com/odwrtw/transmission"
)

var Config struct {
	Addr            string
	User            string
	Password        string
	StartPrivate    bool
	PrivateNotPause bool
}

func initConfig() {
	flag.StringVar(&Config.Addr, "addr", "http://localhost:9091/transmission/rpc", "transmission rpc address")
	flag.StringVar(&Config.User, "user", "", "transmission rpc username")
	flag.StringVar(&Config.Password, "passwd", "", "transmission rpc Password")
	flag.BoolVar(&Config.StartPrivate, "start-private", true, "start all torrents with private tracker")
	flag.BoolVar(&Config.PrivateNotPause, "forever-private", true, "disable auto pause by global seedRatioLimit/seedIdleLimit, for all private torrents")
	flag.Parse()
}

func main() {
	initConfig()

	var err error
	var t *transmission.Client
	{
		conf := transmission.Config{
			Address:  Config.Addr,
			User:     Config.User,
			Password: Config.Password,
		}
		t, err = transmission.New(conf)
		if err != nil {
			panic(fmt.Errorf("transmission.Client create fail. %v", err))
		}
		fmt.Println("connected to", Config.Addr)
	}

	var torrents []*transmission.Torrent
	{
		torrents, err = t.GetTorrents()
		if err != nil {
			panic(fmt.Errorf("get Torrent fail. %v", err))
		}
		fmt.Printf("list %d torrent tasks\n", len(torrents))
	}

	for _, torrent := range torrents {
		if torrent.IsPrivate {
			if Config.StartPrivate && torrent.Status == 0 {
				fmt.Println("start torrent:\t", torrent.Name)
				if err = torrent.StartNow(); err != nil {
					fmt.Println("start fail:", err)
				}
			}
			if Config.PrivateNotPause && (torrent.SeedIdleMode != 2 || torrent.SeedRatioMode != 2) {
				fmt.Println("set private torrent unlimited:\t", torrent.Name)
				err = torrent.Set(transmission.SetTorrentArg{
					SeedIdleMode:  2,
					SeedRatioMode: 2,
				})
				if err != nil {
					fmt.Println("set fail:", err)
				}
			}
		}
	}
}
