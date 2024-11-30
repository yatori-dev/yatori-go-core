package cqie

func (cache *CqieUserCache) GetCookie() string {
	return cache.cookie
}
func (cache *CqieUserCache) SetCookie(cookie string) {
	cache.cookie = cookie
}

func (cache *CqieUserCache) GetVerCode() string {
	return cache.verCode
}
func (cache *CqieUserCache) SetVerCode(verCode string) {
	cache.verCode = verCode
}

func (cache *CqieUserCache) GetAccess_Token() string {
	return cache.access_token
}
func (cache *CqieUserCache) SetAccess_Token(access_token string) {
	cache.access_token = access_token
}

func (cache *CqieUserCache) GetToken() string {
	return cache.token
}
func (cache *CqieUserCache) SetToken(token string) {
	cache.token = token
}

func (cache *CqieUserCache) GetUserId() string {
	return cache.userId
}
func (cache *CqieUserCache) SetUserId(userId string) {
	cache.userId = userId
}

func (cache *CqieUserCache) GetAppId() string {
	return cache.appId
}
func (cache *CqieUserCache) SetAppId(appId string) {
	cache.appId = appId
}

func (cache *CqieUserCache) GetIpaddr() string {
	return cache.ipaddr
}
func (cache *CqieUserCache) SetIpaddr(ipaddr string) {
	cache.ipaddr = ipaddr
}

func (cache *CqieUserCache) GetDeptId() string {
	return cache.deptId
}
func (cache *CqieUserCache) SetDeptId(deptId string) {
	cache.deptId = deptId
}

func (cache *CqieUserCache) GetOrgId() string {
	return cache.orgId
}
func (cache *CqieUserCache) SetOrgId(orgId string) {
	cache.orgId = orgId
}
func (cache *CqieUserCache) GetUserName() string {
	return cache.userName
}
func (cache *CqieUserCache) SetUserName(userName string) {
	cache.userName = userName
}

func (cache *CqieUserCache) GetOrgMajorId() string {
	return cache.orgMajorId
}
func (cache *CqieUserCache) SetOrgMajorId(orgMajorId string) {
	cache.orgMajorId = orgMajorId
}
func (cache *CqieUserCache) GetMobile() string {
	return cache.mobile
}
func (cache *CqieUserCache) SetMobile(mobile string) {
	cache.mobile = mobile
}
func (cache *CqieUserCache) GetStudentId() string {
	return cache.studentId
}
func (cache *CqieUserCache) SetStudentId(studentId string) {
	cache.studentId = studentId
}
