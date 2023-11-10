package bot

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcron"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"log"
	"os/exec"
	"strings"
	"time"
)

func Job() {
	var ctx = gctx.New()
	chatID := int64(-1002009036403)
	token := "6480147492:AAGGuiVN0HYX6Bj6xYQBZE_uwp_On9MvqEY"
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		panic(err)
	}
	//bot.Debug = true
	g.Dump("服务启动，连接机器人成功：" + bot.Self.UserName)
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	ershouJob(ctx, bot, chatID)
	fksFangJob(ctx, bot, chatID)
	flwFangJob(ctx, bot, chatID)
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil { // If we got a message
			/*log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			msg.ReplyToMessageID = update.Message.MessageID

			bot.Send(msg)*/
		}
	}
}

func ershouJob(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) {
	_, err := gcron.AddSingleton(ctx, "0 0 */1 * * *", func(ctx context.Context) {
		response := g.Client().GetVar(ctx, "https://www.flw.ph/plugin.php?id=zimu_fenlei&model=list&ids=2&cid1=0&page=1")
		var res *ErshouRes
		err := response.Scan(&res)
		if err != nil {
			g.Log().Error(ctx, err)
		}

		for _, item := range res.Data.Plists {
			t1 := gtime.New(time.Now())
			t2 := gtime.New(item.Addtime)
			// 只过滤今天新增的数据
			if t2.After(t1.Add(-1*time.Hour)) && t2.Before(t1) {
				// 创建一个包含多张图片的消息组
				var lists = makeMediaGroups(ctx, chatID, item.Imglist, 2)
				var inviteUrl = "\n\n" + "来自 菲律宾同城租房二手交易频道(https://t.me/+J6z60RexzKswOTI9)"
				var msgs = makeMessageGroups(chatID, item.Con+inviteUrl, 4096)
				// 将消息组添加到文本消息中
				for _, mg := range lists {
					_, err = bot.SendMediaGroup(mg)
					if err != nil {
						g.Log().Error(ctx, err)
					}
				}
				for _, m := range msgs {
					_, err = bot.Send(m)
					if err != nil {
						g.Log().Error(ctx, err)
					}
				}
			}

		}
	}, "每隔1小时发送菲龙网二手信息")
	if err != nil {
		g.Log().Error(ctx, err)
	}
}

func fksFangJob(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) {
	_, err := gcron.AddSingleton(ctx, "0 */45 * * * *", func(ctx context.Context) {
		// 要执行的Node.js脚本
		script := `
		fetch("https://h5.flashdelivery.net/customer/api/house/getHouseListByFilter", {
		method: "POST",
		headers: {
			"Content-Type": "application/json", // 根据实际情况设置适当的 Content-Type
			// 如果需要在请求头中添加其他信息，也可以在这里添加
		},
		body: "VBxLc2CQ6RHFLP//dr4BHBS+u0wfYbTEU5JObe+d5VlrtNEeeEA8jcRr2CKfxwzCix9cKB2TAfcSa1c5PmRV8djH0CS54oiqI/9pYV+OEWZatMZj2ih6TZLVWzG/z+eGzVJ8WwxwpF2kncPn1EPtlxo1Agp0/w7pA+6JfqHH4nS8e8iSz6PiQCsZ8NuDDCQxPGsVok4qATsPIy8DcDFAVrADUHmw4cJOpAYLXEl81c1Tbslihu5EksHjSbb42GXH264ceWK1vwO8tS3HaCN1GOa73EDVuVNlIY1xe+f49y8=", // 将数据转换为 JSON 字符串
		})
		.then((response) => {
			if (!response.ok) {
				throw new Error("Network response was not ok");
			}
			return response.text(); // 解析响应数据
		})
		.then((res) => {
			// 在这里处理成功响应的数据
			console.log(res)
		})
		.catch((error) => {
			// 在这里处理请求或解析响应时的错误
			console.error("Request failed:", error);
		});
	`
		// 创建一个命令来执行Node.js脚本
		cmd := exec.Command("node", "-e", script)
		// 执行命令并获取输出结果
		output, err := cmd.Output()
		if err != nil {
			fmt.Println("执行命令错误:", err)
			return
		}
		var res map[string]interface{}
		var stu *ZufangRes
		json.Unmarshal(output, &res)
		gconv.Scan(res, &stu)
		if stu.Code == 1 {
			for _, item := range stu.Result {
				t1 := gtime.New(time.Now())
				t2 := gtime.New(item.CreateTime)
				// 前1小时的数据
				if t2.After(t1.Add(-45*time.Minute)) && t2.Before(t1) {
					// 创建一个包含多张图片的消息组
					var imgs []string
					json.Unmarshal([]byte(item.HouseResourceImages), &imgs)

					var inviteUrl = "\n\n" + "来自 菲律宾同城租房二手交易频道(https://t.me/+J6z60RexzKswOTI9)"
					msg := "\n" + "标题：" + gconv.String(item.HouseTitle) +
						"\n" + "面积：" + gconv.String(item.HouseArea) + "平米" +
						"\n" + "地址：" + gconv.String(item.HouseAddressName) +
						"\n" + "月租：" + gconv.String(item.MonthlyRent) + "披索" +
						"\n" + "联系方式→" + gstr.ReplaceByArray(gconv.String(item.ContactInfo), []string{"{", "", "}", ""}) +
						inviteUrl

					var lists = makeMediaGroups(ctx, chatID, imgs, 2)
					var msgs = makeMessageGroups(chatID, msg, 4096)
					// 将消息组添加到文本消息中
					for _, mg := range lists {
						_, err = bot.SendMediaGroup(mg)
						if err != nil {
							g.Log().Error(ctx, err)
						}
						time.Sleep(1000 * time.Millisecond)
					}
					for _, m := range msgs {
						_, err = bot.Send(m)
						if err != nil {
							g.Log().Error(ctx, err)
						}
						time.Sleep(1000 * time.Millisecond)
					}
				}

			}
		}
	}, "每隔45分钟发送菲快送房源信息")
	if err != nil {
		g.Log().Error(ctx, err)
	}
}

func flwFangJob(ctx context.Context, bot *tgbotapi.BotAPI, chatID int64) {
	_, err := gcron.AddSingleton(ctx, "0 0 */1 * * *", func(ctx context.Context) {
		c := g.Client()
		c.SetHeaderRaw(`
        Referer: https://www.flw.ph/plugin.php?id=fn_house&m=list&class=2
        Cookie: z67S_2132_ulastactivity=b0e05LnyjplEnIRIrFgraZ9fgtz36y%2Bz%2F5hfDLEbEVJzWR6ELchj; z67S_2132_connect_is_bind=0; z67S_2132_lip=103.104.101.122%2C1693990560; z67S_2132_saltkey=lD77b6Fi; z67S_2132_lastvisit=1698113214; CURAD=10; _ga_0VVX8RGWS8=GS1.1.1698116819.10.0.1698116819.0.0.0; _ga=GA1.1.55599681.1692181215; _ga_WF44D8C5GF=GS1.1.1698144111.12.0.1698144111.0.0.0; z67S_2132_sid=adoO0z; z67S_2132_lastact=1698216706%09plugin.php%09; _ga_JGD4T5H0X5=GS1.1.1698216707.3.0.1698216707.0.0.0
        User-Agent: Mozilla/5.0 (iPhone; CPU iPhone OS 13_2_3 like Mac OS X) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/13.0.3 Mobile/15E148 Safari/604.1
    `)
		response := c.GetVar(ctx, "https://www.flw.ph/plugin.php?id=fn_house:Ajax&f=GetAjaxList&page=0&class=2&formhash=e572fadb&publish_type=&keyword=&keywordto=&province=&city=&dist=&price=&vice_class=&room=&order=&house_type=&decoration_type=&orientation=&configure=&tag=&vr_url=")
		var res []*FlwFangItem
		err := response.Scan(&res)
		if err != nil {
			g.Log().Error(ctx, err)
		}

		if res != nil {
			for _, item := range res {
				if item.Dateline == "1小时前" {
					var lists = makeMediaGroups(ctx, chatID, item.Param.Images, 2)
					var inviteUrl = "\n\n" + "来自 菲律宾同城租房二手交易频道(https://t.me/+J6z60RexzKswOTI9)"
					var msg = "\n" + "标题：" + gconv.String(item.Title) +
						"\n" + "说明：" + gconv.String(item.Content) +
						"\n" + "面积：" + gconv.String(item.Square) + "平米" +
						"\n" + "地址：" + gconv.String(item.Community) +
						"\n" + "月租：" + gconv.String(item.Price) + gconv.String(item.PriceText) +
						"\n" + "联系方式：" + "微信（" + gconv.String(item.Param.Wx) + "）" + "飞机（" + gconv.String(item.Param.Zfj) + "）" + "电话（" + gconv.String(item.Mobile) + "）" +
						inviteUrl

					var msgs = makeMessageGroups(chatID, msg, 4096)
					for _, mg := range lists {
						time.Sleep(1000 * time.Millisecond)
						_, err = bot.SendMediaGroup(mg)
						if err != nil {
							g.Log().Error(ctx, err)
						}
					}
					for _, m := range msgs {
						_, err = bot.Send(m)
						if err != nil {
							g.Log().Error(ctx, err)
						}
					}
				}
			}
		}
	}, "每隔1小时发送菲龙网房源信息")
	if err != nil {
		g.Log().Error(ctx, err)
	}
}

// 将消息内容分成多个部分
func splitMessage(message string, maxLength int) []string {
	var parts []string
	for len(message) > maxLength {
		parts = append(parts, message[:maxLength])
		message = message[maxLength:]
	}
	parts = append(parts, message)
	return parts
}

func makeMessageGroups(chatID int64, msg string, maxLength int) []tgbotapi.MessageConfig {
	// 创建一个新的goquery文档
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(msg))
	if err != nil {
		log.Fatal(err)
	}
	// 提取文本内容
	text := doc.Text()
	var msgSlice = splitMessage(text, maxLength)
	var mc []tgbotapi.MessageConfig
	for _, s := range msgSlice {
		mc = append(mc, tgbotapi.NewMessage(chatID, s))
	}
	return mc
}

func splitMedia(images []string, maxCount int) [][]string {
	var parts [][]string
	for len(images) > maxCount {
		parts = append(parts, images[:maxCount])
		images = images[maxCount:]
	}
	parts = append(parts, images)
	return parts
}

func makeMediaGroups(ctx context.Context, chatID int64, images []string, maxCount int) []tgbotapi.MediaGroupConfig {
	var list []interface{}
	var lists []tgbotapi.MediaGroupConfig
	var imgss = splitMedia(images, maxCount)
	for _, imgs := range imgss {
		for _, url := range imgs {
			photoFile := tgbotapi.FileBytes{
				Name:  url,
				Bytes: g.Client().GetBytes(ctx, url),
			}
			list = append(list, tgbotapi.NewInputMediaPhoto(photoFile))
		}
		lists = append(lists, tgbotapi.NewMediaGroup(chatID, list))
	}
	return lists
}

type ErshouRes struct {
	Status int    `json:"status"`
	Msg    string `json:"msg"`
	Data   struct {
		Cat struct {
			Id           int           `json:"id"`
			Uniacid      int           `json:"uniacid"`
			Name         string        `json:"name"`
			Icon         string        `json:"icon"`
			Pid          int           `json:"pid"`
			Price        string        `json:"price"`
			Endtime      int           `json:"endtime"`
			TaocanType   string        `json:"taocan_type"`
			Tag          string        `json:"tag"`
			Isnav        int           `json:"isnav"`
			AppId        string        `json:"app_id"`
			Path         string        `json:"path"`
			Url          string        `json:"url"`
			Status       int           `json:"status"`
			ListTpl      int           `json:"list_tpl"`
			CatKey       string        `json:"cat_key"`
			CatFid       string        `json:"cat_fid"`
			Sort         int           `json:"sort"`
			Onlyadmin    int           `json:"onlyadmin"`
			Onlysetmeal  int           `json:"onlysetmeal"`
			ShareRefresh int           `json:"share_refresh"`
			ShareDig     int           `json:"share_dig"`
			ShareDigDays int           `json:"share_dig_days"`
			Clilds       []interface{} `json:"clilds"`
		} `json:"cat"`
		Plists []struct {
			Id            int         `json:"id"`
			Cid1          int         `json:"cid1"`
			Cid2          int         `json:"cid2"`
			Catname       string      `json:"catname"`
			Uid           int         `json:"uid"`
			Username      string      `json:"username"`
			Mobile        string      `json:"mobile"`
			Con           string      `json:"con"`
			Diycon        interface{} `json:"diycon"`
			Tags          string      `json:"tags"`
			Imglist       []string    `json:"imglist"`
			Display       int         `json:"display"`
			Audit         int         `json:"audit"`
			Confirm       int         `json:"confirm"`
			Endtime       int         `json:"endtime"`
			Toptime       int         `json:"toptime"`
			Refreshtime   int         `json:"refreshtime"`
			Views         int         `json:"views"`
			Shares        int         `json:"shares"`
			Addtime       int         `json:"addtime"`
			Url           string      `json:"url"`
			Source        string      `json:"source"`
			Tid           int         `json:"tid"`
			Opentel       int         `json:"opentel"`
			Offtheshelf   int         `json:"offtheshelf"`
			Avatar        string      `json:"avatar"`
			ListTpl2Views string      `json:"list_tpl2_views"`
			FabuType      string      `json:"fabu_type"`
			Adimg         string      `json:"adimg,omitempty"`
			Type          int         `json:"type,omitempty"`
			Catid         int         `json:"catid,omitempty"`
			Position      int         `json:"position,omitempty"`
			Newtype       string      `json:"newtype,omitempty"`
			Newcatid      string      `json:"newcatid,omitempty"`
			Noapp         int         `json:"noapp,omitempty"`
			Diycolor      string      `json:"diycolor,omitempty"`
		} `json:"plists"`
		Timestamp int `json:"timestamp"`
	} `json:"data"`
}

type ZufangItem struct {
	FeatureTagList              []int   `json:"featureTagList"`
	HouseSourceName             string  `json:"houseSourceName"`
	BathroomCount               int     `json:"bathroomCount"`
	CooperationMode             int     `json:"cooperationMode"`
	Orientation                 int     `json:"orientation"`
	HouseId                     string  `json:"houseId"`
	HouseDetailAddress          string  `json:"houseDetailAddress"`
	HouseStatus                 int     `json:"houseStatus"`
	ContactInfo                 *string `json:"contactInfo"`
	HouseRegion                 int     `json:"houseRegion"`
	HouseType                   int     `json:"houseType"`
	BrokerUserId                *string `json:"brokerUserId"`
	LivingRoomCount             int     `json:"livingRoomCount"`
	MonthlyRent                 int     `json:"monthlyRent"`
	SupportingFacilitiesTagList []int   `json:"supportingFacilitiesTagList"`
	BedroomCount                int     `json:"bedroomCount"`
	CreateTime                  string  `json:"createTime"`
	DeletedFlag                 int     `json:"deletedFlag"`
	HouseArea                   int     `json:"houseArea"`
	HouseTitle                  string  `json:"houseTitle"`
	HouseAddressName            string  `json:"houseAddressName"`
	Decoration                  int     `json:"decoration"`
	HouseResourceImages         string  `json:"houseResourceImages"`
}

type ZufangRes struct {
	Code   int          `json:"code"`
	Msg    string       `json:"msg"`
	Result []ZufangItem `json:"result"`
}

type FlwFangRes []*FlwFangItem

type FlwFangItem struct {
	Id             string      `json:"id"`
	AgentId        string      `json:"agent_id"`
	PublishType    string      `json:"publish_type"`
	Uid            string      `json:"uid"`
	Username       string      `json:"username"`
	Name           string      `json:"name"`
	Mobile         string      `json:"mobile"`
	Class          string      `json:"class"`
	ViceClass      string      `json:"vice_class"`
	Title          string      `json:"title"`
	SmallArea      string      `json:"small_area"`
	Province       string      `json:"province"`
	City           string      `json:"city"`
	Dist           string      `json:"dist"`
	Community      string      `json:"community"`
	Lat            string      `json:"lat"`
	Lng            string      `json:"lng"`
	Price          string      `json:"price"`
	Square         string      `json:"square"`
	VideoUrl       string      `json:"video_url"`
	VrUrl          string      `json:"vr_url"`
	Floor          string      `json:"floor"`
	CountFloor     string      `json:"count_floor"`
	Room           string      `json:"room"`
	Office         string      `json:"office"`
	Guard          string      `json:"guard"`
	Deposit        string      `json:"deposit"`
	HouseType      string      `json:"house_type"`
	DecorationType string      `json:"decoration_type"`
	Years          string      `json:"years"`
	Orientation    string      `json:"orientation"`
	PropertyRight  string      `json:"property_right"`
	ManagementType interface{} `json:"management_type"`
	ShopsType      interface{} `json:"shops_type"`
	Configure      string      `json:"configure"`
	Tag            string      `json:"tag"`
	Content        string      `json:"content"`
	Param          struct {
		Wx           string      `json:"wx"`
		Zfj          string      `json:"zfj"`
		Mastermobile string      `json:"mastermobile"`
		PriceTime    int         `json:"price_time"`
		Images       []string    `json:"images"`
		Cover        string      `json:"cover"`
		TagList      interface{} `json:"tag_list"`
		Content      string      `json:"content"`
	} `json:"param"`
	Display         string `json:"display"`
	Hot             string `json:"hot"`
	PaymentState    string `json:"payment_state"`
	Deal            string `json:"deal"`
	Click           string `json:"click"`
	Dateline        string `json:"dateline"`
	EditDateline    string `json:"edit_dateline"`
	Updateline      string `json:"updateline"`
	Topdateline     string `json:"topdateline"`
	IsFree          string `json:"is_free"`
	OTilte          string `json:"o_tilte"`
	OContent        string `json:"o_content"`
	Url             string `json:"url"`
	Ftitle          string `json:"ftitle"`
	Huxing          string `json:"huxing"`
	ProvinceText    string `json:"province_text"`
	DisplayText     string `json:"display_text"`
	OverdueText     string `json:"overdue_text"`
	SeeUpdateline   string `json:"see_updateline"`
	PublishTypeText string `json:"publish_type_text"`
	ReturnTop       int    `json:"return_top"`
	ViceClassText   string `json:"vice_class_text"`
	PriceText       string `json:"price_text"`
	Refresh         int    `json:"refresh"`
	Top             int    `json:"top"`
}
