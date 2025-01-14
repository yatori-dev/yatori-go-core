package service

import (
	"github.com/robfig/cron/v3"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/entity"
	"github.com/yatori-dev/yatori-go-core/api/gongxue/global"
	gongxueyun_service2 "github.com/yatori-dev/yatori-go-core/api/gongxue/service/gongxueyun_service"
	"time"
)

type AppService struct {
	users []entity.SignEntity
	cron  *cron.Cron
}

func NewAppService() *AppService {
	return &AppService{
		cron: cron.New(),
	}
}

func (svc *AppService) Init() {
	if !global.Config.Account.Promptly {
		svc.scheduleTasks()
		svc.cron.Start()
	} else {
		svc.StartGongxueYun("sign")
	}
	//svc.StartTestCX()
	select {}
}

func (svc *AppService) scheduleTasks() {
	global.Log.Info("Scheduling tasks...")

	svc.addCronTask("30 8 * * *", "每天早上8点30签到", "sign")
	svc.addCronTask("30 17 * * *", "每天晚上5点30签到", "sign")
	//svc.addCronTask("35 23 * * *", "每天早上8点30签到", "sign")
	//svc.addCronTask("35 14 * * *", "每天晚上6点30签到", "sign")
	//svc.addCronTask("0 10 * * 5", "每周周五早上10点签到", "week")
	//svc.cron.AddFunc("0 10 ? * 1L", func() {
	//	if isLastWeek(time.Now()) {
	//		global.Log.Info("Running task: 每月最后一周的周一早上10点签到")
	//		svc.StartGongxueYun("month")
	//		global.Log.Info("Task finished!")
	//	}
	//})
}

func (svc *AppService) addCronTask(schedule, logMessage, taskType string) {
	svc.cron.AddFunc(schedule, func() {
		global.Log.Infof("Running task: %s", logMessage)
		svc.StartGongxueYun(taskType)
		global.Log.Info("Task finished!")
	})
}

func (svc *AppService) StartGongxueYun(taskType string) {
	svc.users = gongxueyun_service2.LoadUsers()
	global.Log.Info("Starting Gongxueyun module...")
	for _, user := range svc.users {
		CreateMoguDing(user).Run(taskType)
	}
}

func CreateMoguDing(user entity.SignEntity) *gongxueyun_service2.MoguDing {
	return &gongxueyun_service2.MoguDing{
		ID:          user.ID,
		PhoneNumber: user.Username,
		Password:    user.Password,
		Email:       user.Email,
		Sign: gongxueyun_service2.SignInfo{
			City:      user.City,
			Area:      user.Area,
			Address:   user.Address,
			Country:   user.Country,
			Province:  user.Province,
			Latitude:  user.Latitude,
			Longitude: user.Longitude,
		},
	}
}

func isLastWeek(t time.Time) bool {
	_, week := t.ISOWeek()
	nextMonday := t.AddDate(0, 0, 7-int(t.Weekday()))
	nextMonthWeek, _ := nextMonday.ISOWeek()
	return week != nextMonthWeek
}
