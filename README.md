# GoSwissArmy
A collection of handy utilities developed by Golang

## Transmission 辅种工具 tr

我会使用 Transmission 下载 PT，也会用它下载一些公网 BT 资源。对于公网资源，全局设置了 \[停止做种超时时间、停止做种分享率上限\]，不然小水管受不了。

**但对于 PT 资源，我希望持续做种。** 所以该工具会自动扫描 tr 当中的 PT 资源，禁用任务级别的 \[停止做种\] 开关。

```
  -addr string
    	transmission rpc address (default "http://localhost:9091/transmission/rpc")
  -forever-private
    	disable auto pause by global seedRatioLimit/seedIdleLimit, for all private torrents (default true)
  -passwd string
    	transmission rpc Password
  -start-private
    	start all torrents with private tracker (default true)
  -user string
    	transmission rpc username
```

1. 通过 addr、user、passwd 连接到 tr
2. start-private 启用时，将所有 private 任务启动
3. forever-private 启用时，为所有 private 任务禁用任务级别的 \[停止做种\] 开关

## IPTV 频道工厂 iptv

通过 IPTV 看电视主打一个省钱，网上有很多 m3u 清单可以获取到世界各地的频道清单，对这些清单的合并、去重、清理无效有很多实用工具。

但是， **频道能连接并不意味着可以流畅播放。** 该工具在聚合多个 m3u 的同时，探测每个频道的下载速度，筛选在本地网络下能流畅播放的频道。

```
  -coroutine int
    	coroutine to test lists (default 4)
  -lists string
    	m3u8 lists split by | (default "https://iptv-org.github.io/iptv/languages/zho.m3u")
  -output string
    	output file path (default "./iptv.m3u8")
  -tolerate int
    	allowed block percents of the stream (default 20)
```

下载速度 > 视频流生成速度 * (100-tolerate)%，认为该频道可以流畅播放。多个频道同时检测，并发度由 coroutine 控制。

## 物联网远程开关机 mqttiot

智能家居让开灯开空调都变成了一句话的事，如果能 **把家里电脑也接入米家，语音开关机** 那就方便了，配合自动化可以打造一些智能场景，出门在外也可以远程遥控。

该方案同时也支持天猫精灵、小度等平台。

```
  -halt
    	enable halt system by mqtt
  -mqtt-addr string
    	MQTT Address (default "wss://bemfa.com:9504/wss")
  -mqtt-cli-id string
    	MQTT ClientID (default "xxx")
  -mqtt-topic string
    	MQTT topic (default "hzyspc001")
  -wol-addr string
    	IP address for wol (default "192.168.0.255:9")
  -wol-mac string
    	MAC address for wol
```

1. 使用 mqtt-addr、mqtt-cli-id 连接到 MQTT，并持续监听 mqtt-topic 中的消息
2. 接受到 "on" 消息并且 wol-mac 非空时，触发远程开机：WOL 唤醒 wol-addr、wol-mac
3. 接受到 "off" 消息并且 halt 启用时，触发 Windows 系统本地关机

巴法云提供了免费的 MQTT 服务并支持接入米家 IOT：<https://cloud.bemfa.com/>

这样，在 NAS 中持续运行此工具，就能接受开机指令，远程唤醒 PC。PC 开机后运行此工具，收到关机指令后，就能自动关机 😄
