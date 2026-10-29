package data

type LoginData struct {
	AuthType    int    `json:"authType"`
	ExpiredTime int64  `json:"expiredTime"`
	Gender      int    `json:"gender"`
	HeadImg     string `json:"headImg"`
	IsAccount   int    `json:"isAccount"`
	MoguNo      string `json:"moguNo"`
	MsgConfig   struct {
	} `json:"msgConfig"`
	NeteaseImId string `json:"neteaseImId"`
	NikeName    string `json:"nikeName"`
	OrgJson     struct {
		ClassId       string `json:"classId"`
		ClassName     string `json:"className"`
		DepId         string `json:"depId"`
		DepName       string `json:"depName"`
		Grade         string `json:"grade"`
		MajorField    string `json:"majorField"`
		MajorId       string `json:"majorId"`
		MajorName     string `json:"majorName"`
		SchoolId      string `json:"schoolId"`
		SchoolName    string `json:"schoolName"`
		SnowFlakeId   string `json:"snowFlakeId"`
		StudentId     string `json:"studentId"`
		StudentNumber string `json:"studentNumber"`
		UserName      string `json:"userName"`
	} `json:"orgJson"`
	Phone     string `json:"phone"`
	RoleGroup []struct {
		CurrPage   int    `json:"currPage"`
		OrderBy    string `json:"orderBy"`
		PageSize   int    `json:"pageSize"`
		RoleId     string `json:"roleId"`
		RoleKey    string `json:"roleKey"`
		RoleLevel  string `json:"roleLevel"`
		RoleName   string `json:"roleName"`
		Sort       string `json:"sort"`
		TotalCount int    `json:"totalCount"`
		TotalPage  int    `json:"totalPage"`
	} `json:"roleGroup"`
	RoleId    string `json:"roleId"`
	RoleKey   string `json:"roleKey"`
	RoleLevel string `json:"roleLevel"`
	RoleName  string `json:"roleName"`
	Token     string `json:"token"`
	UserId    string `json:"userId"`
	UserType  string `json:"userType"`
}
