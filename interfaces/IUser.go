package interfaces

/*
userName: 用户名称
phone:电话号码
address: 地址
*/
// 用户账号接口
type IUser interface {
	Login() map[string]any     //登录
	UserInfo() map[string]any  //用户个人信息
	CacheData() map[string]any //账号缓存信息
	CourseList() []ICourse     //课程列表
}

// 课程接口
type ICourse interface {
	CourseName() string //课程名称
	TaskList() []ITask  //任务拉取
}

// 最小任务单位
type ITask interface {
	Type() string                                      //任务类型
	DataMap() map[string]interface{}                   //存储的相关数据
	Start(callback func(result map[string]any)) string //执行任务
	Stop(callback func(result map[string]any)) string  //暂停任务
	Kill(callback func(result map[string]any)) string  //直接取消任务
}
