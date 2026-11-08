package driverutil

import (
	"testing"
	"time"
)

func TestCourseDayEndKeepsEndDateInclusive(t *testing.T) {
	endDate, err := time.ParseInLocation("2006-01-02", "2026-11-08", time.Local)
	if err != nil {
		t.Fatalf("解析日期失败: %v", err)
	}

	endOfDay := CourseDayEnd(endDate)
	if endOfDay.Year() != 2026 || endOfDay.Month() != time.November || endOfDay.Day() != 8 {
		t.Fatalf("结束时间被改到了别的日期: %v", endOfDay)
	}

	lastMoment := time.Date(2026, 11, 8, 23, 59, 59, 0, time.Local)
	if lastMoment.After(endOfDay) {
		t.Fatalf("到期当天 23:59 不应判定为已结束: endOfDay=%v", endOfDay)
	}

	nextDay := time.Date(2026, 11, 9, 0, 0, 0, 0, time.Local)
	if !nextDay.After(endOfDay) {
		t.Fatalf("到期次日应判定为已结束: endOfDay=%v", endOfDay)
	}
}

func TestCourseDayStartKeepsStartDateAvailable(t *testing.T) {
	startDate, err := time.ParseInLocation("2006-01-02", "2026-11-08", time.Local)
	if err != nil {
		t.Fatalf("解析日期失败: %v", err)
	}

	startOfDay := CourseDayStart(startDate)
	firstMoment := time.Date(2026, 11, 8, 0, 0, 0, 0, time.Local)
	if firstMoment.Before(startOfDay) {
		t.Fatalf("开课当天 00:00 不应判定为未开始: startOfDay=%v", startOfDay)
	}

	beforeStart := time.Date(2026, 11, 7, 23, 59, 59, 0, time.Local)
	if !beforeStart.Before(startOfDay) {
		t.Fatalf("开课前一天 23:59 应判定为未开始: startOfDay=%v", startOfDay)
	}
}
