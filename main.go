package main

import (
	"douyin-grab/constv"
	"douyin-grab/grab"
	"douyin-grab/nmid"
	"douyin-grab/pkg/logger"
	"douyin-grab/wsocket"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"
	"github.com/urfave/cli"
)

const (
	VERSION = `0.0.2`
)

func main() {
	app := cli.NewApp()
	app.Name = `Douyin Grab`
	app.Version = showVersion()
	app.Before = func(ctx *cli.Context) error {
		return nil
	}

	showBanner()
	godotenv.Load("./.env")
	logger.Init("")

	// var live_room_url, wss_url string
	var live_room_id string
	app.Flags = []cli.Flag{
		// cli.StringFlag{
		// 	Name:        "live_room_url, lrurl",
		// 	Usage:       "live room url",
		// 	Destination: &live_room_url,
		// },
		// cli.StringFlag{
		// 	Name:        "wss_url, wssurl",
		// 	Usage:       "live room wws url",
		// 	Destination: &wss_url,
		// },
		cli.StringFlag{
			Name:        "live_room_id, lrid",
			Usage:       "live room id",
			Destination: &live_room_id,
		},
	}

	var err error
	app.Action = func(ctx *cli.Context) error {
		// if len(live_room_url) == 0 {
		// 	live_room_url = constv.DEFAULTLIVEROOMURL //默认直播间url
		// }
		// logger.Info("live room url: %s", live_room_url)

		// if len(wss_url) == 0 {
		// 	wss_url = constv.DEFAULTLIVEWSSURL //默认直播间wss_url
		// }
		// logger.Info("live room wss_url: %s", wss_url)

		live_room_url := constv.DEFAULTLIVEROOMURL //默认直播间url
		wss_url := constv.DEFAULTLIVEWSSURL        //默认直播间wss_url

		if len(live_room_id) == 0 {
			live_room_id = constv.DEFAULTLIVEROOMID //默认直播间id
		}
		logger.Info("live room id: %s", live_room_id)

		if len(live_room_id) > 0 {
			live_room_url = fmt.Sprintf("%s/%s", constv.DOUYIORIGIN, live_room_id)
			logger.Info("live room url: %s", live_room_url)
			//wssUrl := "wss://webcast5-ws-web-hl.douyin.com/webcast/im/push/v2/?app_name=douyin_web&version_code=180800&webcast_sdk_version=1.0.8&update_version_code=1.0.8&compress=gzip&device_platform=web&cookie_enabled=true&screen_width=1920&screen_height=1200&browser_language=zh-CN&browser_platform=Win32&browser_name=Mozilla&browser_version=5.0%20(Windows%20NT%2010.0;%20Win64;%20x64)%20AppleWebKit/537.36%20(KHTML,%20like%20Gecko)%20Chrome/116.0.0.0%20Safari/537.36&browser_online=true&tz_name=Asia/Shanghai&cursor=h-1_t-1696830455713_r-1_d-1_u-1&internal_ext=internal_src:dim|wss_push_room_id:7287821924091054848|wss_push_did:7276281074124408375|dim_log_id:20231009134735DA37891DD2868B18133A|first_req_ms:1696830455641|fetch_time:1696830455713|seq:1|wss_info:0-1696830455713-0-0|wrds_kvs:WebcastRoomRankMessage-1696830341446566674_WebcastRoomStatsMessage-1696830449393067841_LotteryInfoSyncData-1696829719563436524_HighlightContainerSyncData-2&host=https://live.douyin.com&aid=6383&live_id=1&did_rule=3&endpoint=live_pc&support_wrds=1&user_unique_id=7276281074124408375&im_path=/webcast/im/fetch/&identity=audience&room_id=7287821924091054848&heartbeatDuration=0&signature=Wko/DCJQEn3gF+2c"
			//wssUrl := constv.DEFAULTLIVEWSSURL
			wssUrl, err := grab.GetWssUrl(live_room_url)
			if err != nil {
				logger.Error("GetWssUrl error: %v", err.Error())
			}
			logger.Info("get wss url %s", wssUrl)
			if nil == err {
				wss_url = wssUrl
			}
		}

		//获取直播间信息
		_, ttwid := grab.FetchLiveRoomInfo(live_room_url)

		//与直播间进行websocket通信，获取评论数据
		header := http.Header{}
		header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36") // 设置User-Agent头
		header.Set("Origin", constv.DOUYIORIGIN)
		cookie := &http.Cookie{
			Name:  "ttwid",
			Value: ttwid,
		}
		header.Add("Cookie", cookie.String())
		wsclient := wsocket.NewWSClient().SetRequestInfo(wss_url, header)
		wsclient.ConnWSServer(ttwid)
		wsclient.RunWSClient()

		//worker服务
		go nmid.RunWorker()

		return nil
	}

	err = app.Run(os.Args)
	if err != nil {
		panic(err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	os.Exit(0)
}

func showBanner() {
	println(`
	 _| _      . _    _  _ _  _|
	(_|(_)|_|\/|| |  (_|| (_|(_|
			 /        _|      `)
}

func showVersion() string {
	bannerData := `
	 _| _      . _    _  _ _  _|
	(_|(_)|_|\/|| |  (_|| (_|(_|
			 /        _|      `
	return bannerData + "\n" + VERSION
}
