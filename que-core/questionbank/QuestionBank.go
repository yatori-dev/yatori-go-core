package questionbank

import (
	"crypto/md5"
	"errors"
	"fmt"
	log2 "github.com/yatori-dev/yatori-go-core/utils/log"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"time"
)

type Question struct {
	gorm.Model
	Md5     string `gorm:"column:md5"`              //题目MD5值，注意，是（题目类型+题目内容）的编码的MD5值
	Type    string `gorm:"column:type"`             //题目类型
	Content string `gorm:"column:content"`          //题目内容
	Options string `gorm:"column:options"`          //选项（一般选择题才会有），存储为Json
	Answers string `gorm:"column:answer;type:TEXT"` // 答案，存储为 JSON
}

//// 在保存之前将 Answer 转换为 JSON 字符串
//func (q *Question) BeforeSave(tx *gorm.DB) error {
//	data, err := json.Marshal(q.Answers)
//	if err != nil {
//		return err
//	}
//	q.Answer = string(data)
//	return nil
//}

//// 在查询之后将 JSON 字符串转换回 Answer 数组
//func (q *Question) AfterFind(tx *gorm.DB) error {
//	var data []string
//	if err := json.Unmarshal([]byte(q.Content), &data); err != nil {
//		return err
//	}
//	q.Answer = data
//	return nil
//}

// 题库缓存初始化
func QuestionBackInit() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("questionbank.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Question{})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB() //数据库连接池
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	return db, nil
}

// 插入题库
func (question *Question) Insert(db *gorm.DB) error {
	if err := db.Create(&question).Error; err != nil {
		return errors.New("插入数据失败: " + err.Error())
	}
	log2.Print(log2.DEBUG, "插入数据成功")
	return nil
}

// 如果没有则插入题库
func (question *Question) InsertIfNot(db *gorm.DB) error {
	//检查是否合法题目
	checkErr := CheckQue(question)
	if checkErr != nil {
		return checkErr
	}
	selectQs := question.SelectsForTypeAndContent(db)
	if len(selectQs) > 0 {
		return nil
	}
	if question.Md5 == "" {
		question.Md5 = fmt.Sprintf("%x", md5.Sum([]byte(fmt.Sprintf("%s-%s", question.Type, question.Content))))
	}
	// 插入题目
	err := question.Insert(db)
	if err != nil {
		return err
	}
	return nil
}

// 根据题目类型和内容查询题目
func (question *Question) SelectsForTypeAndContent(db *gorm.DB) []Question {
	var questions []Question
	if err := db.Where("type = ? AND content = ?", question.Type, question.Content).Find(&questions).Error; err != nil {
		log.Fatalf("查询数据失败: %v", err)
	}
	return questions
}

// 根据题目MD5查询
func (question *Question) SelectsForMd5(db *gorm.DB) []Question {
	var questions []Question
	if err := db.Where("md5 = ?", question.Md5).Find(&questions).Error; err != nil {
		log.Fatalf("查询数据失败: %v", err)
	}
	return questions
}

// 直接通过题目找答案返回
func (question *Question) SelectAnswer(db *gorm.DB) []string {

	return nil
}

// 根据题目类型和内容更新题目
func (question *Question) UpdateAnswerForTypeAndContent(db *gorm.DB) error {
	if err := db.Where("type = ? AND content = ?", question.Type, question.Content).Updates(&question).Error; err != nil {
		return err
	}
	return nil
}

// 根据题目类型和内容删除题目
func (question *Question) DeleteForTypeAndContent(db *gorm.DB) error {
	if err := db.Where("type = ? AND content = ?", question.Type, question.Content).Delete(&Question{}).Error; err != nil {
		return err
	}
	return nil
}

// 检验Question合法性
func CheckQue(question *Question) error {
	//检验数据合法性
	if question.Type == "" {
		return errors.New("Not Found Question Type")
	}
	if question.Content == "" {
		return errors.New("Not Found Question Content")
	}
	return nil
}
