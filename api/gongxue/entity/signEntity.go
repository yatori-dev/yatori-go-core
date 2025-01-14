package entity

// Sign 签到实体
type SignEntity struct {
	ID        int    `gorm:"primaryKey;autoIncrement;column:id"` // 主键，自增
	Username  string `gorm:"unique;not null;column:username"`    // 唯一，非空
	Password  string `gorm:"not null;column:password"`           // 非空
	Country   string `gorm:"column:country"`
	Province  string `gorm:"column:province"`
	City      string `gorm:"column:city"`
	Area      string `gorm:"column:area"`
	Latitude  string `gorm:"column:latitude"`
	Longitude string `gorm:"column:longitude"`
	Email     string `gorm:"unique;column:email"` // 唯一
	Address   string `gorm:"column:address"`
	Type      int    `gorm:"not null;default:0;column:type"`  // 非空，默认值
	State     int    `gorm:"not null;default:0;column:state"` // 非空，默认值
	Token     string `gorm:"column:token"`
}

// TableName 指定表名
func (SignEntity) TableName() string {
	return "sign"
}
