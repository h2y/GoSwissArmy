package main

import (
	"flag"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mdlayher/wol"
	"log"
	"net"
	"os"
	"os/exec"
	"time"
)

var Config struct {
	MqttAddr     string
	MqttClientID string
	MqttTopic    string
	Halt         bool
	WolAddr      string

	wol    bool
	wolMac net.HardwareAddr
}

func initConfig() {
	wolMac := flag.String("wol-mac", "", "MAC address for wol")
	flag.StringVar(&Config.WolAddr, "wol-addr", "192.168.0.255:9", "IP address for wol")
	flag.StringVar(&Config.MqttClientID, "mqtt-cli-id", "xxx", "MQTT ClientID")
	flag.StringVar(&Config.MqttAddr, "mqtt-addr", "wss://bemfa.com:9504/wss", "MQTT Address")
	flag.StringVar(&Config.MqttTopic, "mqtt-topic", "hzyspc001", "MQTT topic")
	flag.BoolVar(&Config.Halt, "halt", false, "enable halt system by mqtt (Windows OS only)")
	flag.Parse()

	var err error
	if len(*wolMac) > 0 {
		Config.wol = true
		Config.wolMac, err = net.ParseMAC(*wolMac)
		if err != nil {
			log.Panicln("ParseMAC fail", err)
		}
	}

	log.Printf("start with config: %+v\n", Config)
}

func mqttMsgHandler(_ mqtt.Client, msg mqtt.Message) {
	payout := string(msg.Payload())
	log.Printf("msg: %v\n", payout)

	if payout == "on" && Config.wol {
		client, err := wol.NewClient()
		if err != nil {
			log.Fatal("wol.NewClient fail", err)
		}
		defer client.Close()
		err = client.Wake(Config.WolAddr, Config.wolMac)
		if err != nil {
			log.Printf("fail to wol: %v\n", err)
		}
	} else if payout == "off" && Config.Halt {
		cmd := exec.Command("shutdown", "/s", "/f")
		err := cmd.Run()
		log.Fatal("shutdown ret:", err)
	} else {
		log.Println("skip")
	}
}

func main() {
	initConfig()

	mqconf := mqtt.NewClientOptions()
	mqconf.AddBroker(Config.MqttAddr)
	mqconf.SetClientID(Config.MqttClientID)
	mqconf.SetCleanSession(true).SetAutoReconnect(true)
	mqconf.SetOnConnectHandler(func(mqc mqtt.Client) {
		t := mqc.Subscribe(Config.MqttTopic, byte(1), mqttMsgHandler)
		if !t.Wait() || t.Error() != nil {
			log.Printf("ERROR: Subscribe fail %v\n", t.Error())
			time.Sleep(time.Second * 5)
			os.Exit(1)
		}
		log.Println("mqtt Connect")
	})
	mqc := mqtt.NewClient(mqconf)

	t := mqc.Connect()
	for !t.Wait() || t.Error() != nil {
		log.Printf("ERROR: Connect fail %v\n", t.Error())
		time.Sleep(time.Second * 5)
		t = mqc.Connect()
	}

	select {}
}
