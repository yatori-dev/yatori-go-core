package questionback

import (
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"time"
)

type Question struct {
	gorm.Model
	Type    string `gorm:"column:type"`             //题目类型
	Content string `gorm:"column:content"`          //题目内容
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
	db, err := gorm.Open(sqlite.Open("questionback.db"), &gorm.Config{})
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

// 插入缓存题库
func (question *Question) QuestionBackInsert(db *gorm.DB) error {
	if err := db.Create(&question).Error; err != nil {
		return errors.New("插入数据失败: " + err.Error())
	}
	log.Println("插入数据成功")
	return nil
}

// 根据题目类型和内容查询题目
func (question *Question) QuestionBackSelectsForTypeAndContent(db *gorm.DB) []Question {
	var questions []Question
	if err := db.Where("type = ? AND content = ?", question.Type, question.Content).Find(&questions).Error; err != nil {
		log.Fatalf("查询数据失败: %v", err)
	}
	return questions
}

// 根据题目类型和内容更新题目
func (question *Question) QuestionBackUpdateAnswerForTypeAndContent(db *gorm.DB) error {
	if err := db.Where("type = ? AND content = ?", question.Type, question.Content).Updates(&question).Error; err != nil {
		return err
	}
	return nil
}

// 根据题目类型和内容删除题目
func (question *Question) QuestionBackDeleteForTypeAndContent(db *gorm.DB) error {
	if err := db.Where("type = ? AND content = ?", question.Type, question.Content).Delete(&Question{}).Error; err != nil {
		return err
	}
	return nil
}
