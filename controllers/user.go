package controllers

import (
	"fmt"
	"iuv520/daily-reminder/orm"
	"iuv520/daily-reminder/services"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
)

//User 用户相关
type User struct {
}

//Post 基础post
func (u *User) Post(ctx iris.Context) *JSONResponse {
	req := orm.UserContent{}
	err := ctx.ReadJSON(&req)
	if err != nil {
		log.Printf("%v", err)
	}

	uuid := uuid.New().String()

	modelUser := orm.UserInfo{}
	modelUser.Name = uuid
	modelUser.Content = req
	engine := orm.GetEngine()

	_, err = engine.Insert(modelUser)
	if err != nil {
		log.Printf("%v", err)
	}
	return &JSONResponse{
		Code: 200,
		Msg:  "success",
		Data: uuid,
	}
}

//GetLotteryBy 获取彩票中奖情况
func (u *User) GetLotteryBy(userID string) *JSONResponse {
	calculator := &services.Calculator{}

	res, err := calculator.Run(userID)
	if err != nil {
		return &JSONResponse{
			Code: 500,
			Msg:  err.Error(),
			Data: nil,
		}
	}
	return &JSONResponse{
		Code: 200,
		Msg:  "success",
		Data: res,
	}
}

//GetWeatherBy 获取天气情况
func (u *User) GetWeatherBy(userID string) *JSONResponse {
	weather := &services.Weather{}

	res, err := weather.Run(userID)
	if err != nil {
		return &JSONResponse{
			Code: 500,
			Msg:  err.Error(),
			Data: nil,
		}
	}
	return &JSONResponse{
		Code: 200,
		Msg:  "success",
		Data: res,
	}
}

//GetHotsearch 微博热搜
func (u *User) GetHotsearch() *JSONResponse {
	engine := orm.GetEngine()
	hotSearchList := make([]orm.HotSearch, 0)
	err := engine.Cols("Content").Where("date_format(created_at,'%Y-%m-%d') = ?", time.Now().Format("2006-01-02")).OrderBy("created_at desc, score desc").Limit(10).Find(&hotSearchList)
	res := make([]string, 0)
	for _, v := range hotSearchList {
		res = append(res, v.Content)
	}
	if err != nil {
		return &JSONResponse{
			Code: 500,
			Msg:  err.Error(),
			Data: nil,
		}
	}
	return &JSONResponse{
		Code: 200,
		Msg:  "success",
		Data: res,
	}
}

//GetReportBy 总计文字
func (u *User) GetReportBy(userID string) *JSONResponse {
	lottery := u.GetLotteryBy(userID).Data.(map[string]int)
	weather := u.GetWeatherBy(userID).Data.(*orm.Weather)
	hotSearch := u.GetHotsearch().Data.([]string)

	finalWord := make([]string, 0)

	finalWord = append(finalWord, fmt.Sprintf("早安啊。今天是%s，让我们来看一下乱七八糟的事情吧。", time.Now().Format("2006年01月02日")))

	// 彩票系列
	if len(lottery) > 0 {
		eachLottery := make([]string, 0)
		lotteryTotal := 0
		for k, v := range lottery {
			if v == -1 {
				word := fmt.Sprintf("卧槽昨晚的%s你中大奖了！赶紧自己瞅一眼！", u.translateLottery(k))
				return &JSONResponse{
					Code: 200,
					Msg:  "success",
					Data: word,
				}
			}

			lotteryTotal += v
			if v != 0 {
				eachLottery = append(eachLottery, fmt.Sprintf("%s中了%d元，", u.translateLottery(k), v))
			}
		}
		if lotteryTotal == 0 {
			finalWord = append(finalWord, fmt.Sprintf("有关彩票，又没中奖，习惯就好。"))
		} else {
			finalWord = append(finalWord, fmt.Sprintf("昨晚的彩票你一共中了%d元，其中，%s。记得兑奖哦。你也只配中这点了吧。", lotteryTotal, strings.Join(eachLottery, "，")))
		}
	}

	// 天气系列

	if weather != nil {
		finalWord = append(finalWord, fmt.Sprintf("今天%s的气温是%.0f摄氏度，%s%s。空气质量指数%d。", weather.City, weather.Temperature, weather.Direction, weather.Power, weather.Aqi))
	}
	// 热搜系列
	finalHotWord := "此刻微博热搜：" + strings.Join(hotSearch, "；") + "。"

	finalWord = append(finalWord, finalHotWord)

	return &JSONResponse{
		Code: 200,
		Msg:  "success",
		Data: strings.Join(finalWord, "\n"),
	}

}

func (u *User) translateLottery(name string) string {
	switch name {
	case "ssq":
		return "双色球"
	case "fcsd":
		return "福彩3D"
	case "dlt":
		return "大乐透"
	default:
		return "其它"
	}
}
