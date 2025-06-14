package interfaces

/*IUser
userName: 用户名称
phone:电话号码
address: 地址
*/
// 用户账号接口
type IUser interface {
	Login() (map[string]any, error)     //登录
	UserInfo() (map[string]any, error)  //用户个人信息
	CacheData() (map[string]any, error) //账号缓存信息
	CourseList() ([]ICourse, error)     //课程列表
}

// ICourse 课程接口
type ICourse interface {
	GetCourseID() string
	GetCourseName() string //课程名称
	TaskList() []ITask     //任务拉取
}

// ITask 最小任务单位
type ITask interface {
	Type() (string, error)                                      //任务类型
	DataMap() (map[string]interface{}, error)                   //存储的相关数据
	Start(callback func(result map[string]any)) (string, error) //执行任务
	Stop(callback func(result map[string]any)) (string, error)  //暂停任务
	Kill(callback func(result map[string]any)) (string, error)  //直接取消任务
}
