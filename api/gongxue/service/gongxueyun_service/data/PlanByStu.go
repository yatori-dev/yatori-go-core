package data

type PlanByStuData struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data []struct {
		TotalCount     int         `json:"totalCount"`
		PageSize       int         `json:"pageSize"`
		TotalPage      int         `json:"totalPage"`
		CurrPage       int         `json:"currPage"`
		IsDeleted      interface{} `json:"isDeleted"`
		CreateBy       interface{} `json:"createBy"`
		ModifiedBy     interface{} `json:"modifiedBy"`
		CreateTime     interface{} `json:"createTime"`
		ModifiedTime   interface{} `json:"modifiedTime"`
		CreateByName   interface{} `json:"createByName"`
		ModifiedByName interface{} `json:"modifiedByName"`
		OrderBy        string      `json:"orderBy"`
		Sort           string      `json:"sort"`
		PlanId         string      `json:"planId"`
		SchoolId       interface{} `json:"schoolId"`
		DepId          interface{} `json:"depId"`
		DepName        interface{} `json:"depName"`
		PlanName       string      `json:"planName"`
		PlanNumber     interface{} `json:"planNumber"`
		PlanLevel      interface{} `json:"planLevel"`
		PlanGrades     interface{} `json:"planGrades"`
		Type           string      `json:"type"`
		StartTime      string      `json:"startTime"`
		EndTime        string      `json:"endTime"`
		Subsidy        interface{} `json:"subsidy"`
		Description    interface{} `json:"description"`
		PlanState      interface{} `json:"planState"`
		Backup         interface{} `json:"backup"`
		IsSign         int         `json:"isSign"`
		IsAuto         string      `json:"isAuto"`
		BatchId        string      `json:"batchId"`
		PracticeStus   interface{} `json:"practiceStus"`
		SnowFlakeId    interface{} `json:"snowFlakeId"`
		BatchName      string      `json:"batchName"`
		PlanPaper      struct {
			IsDeleted          int    `json:"isDeleted"`
			CreateTime         string `json:"createTime"`
			PlanPaperId        string `json:"planPaperId"`
			PlanId             string `json:"planId"`
			DayPaperNum        int    `json:"dayPaperNum"`
			WeekPaperNum       int    `json:"weekPaperNum"`
			MonthPaperNum      int    `json:"monthPaperNum"`
			SummaryPaperNum    int    `json:"summaryPaperNum"`
			WeekReportCount    int    `json:"weekReportCount"`
			PaperReportCount   int    `json:"paperReportCount"`
			MonthReportCount   int    `json:"monthReportCount"`
			SummaryReportCount int    `json:"summaryReportCount"`
			MaxDayNum          int    `json:"maxDayNum"`
			MaxWeekNum         int    `json:"maxWeekNum"`
			MaxMonthNum        int    `json:"maxMonthNum"`
			MaxSummaryNum      int    `json:"maxSummaryNum"`
			SnowFlakeId        int    `json:"snowFlakeId"`
			DayPaper           bool   `json:"dayPaper"`
			WeekPaper          bool   `json:"weekPaper"`
			MonthPaper         bool   `json:"monthPaper"`
			SummaryPaper       bool   `json:"summaryPaper"`
		} `json:"planPaper"`
		PlanPaperMap             interface{} `json:"planPaperMap"`
		Attachments              interface{} `json:"attachments"`
		PlanMajors               interface{} `json:"planMajors"`
		PlanClasses              interface{} `json:"planClasses"`
		PlanAppraiseItem         interface{} `json:"planAppraiseItem"`
		PlanAppraiseItemDtos     interface{} `json:"planAppraiseItemDtos"`
		PlanAppraiseItemEntities interface{} `json:"planAppraiseItemEntities"`
		MajorNames               interface{} `json:"majorNames"`
		CreateName               string      `json:"createName"`
		AttachmentNum            int         `json:"attachmentNum"`
		PlanIds                  interface{} `json:"planIds"`
		SignCount                interface{} `json:"signCount"`
		AuditState               interface{} `json:"auditState"`
		MajorTeacher             interface{} `json:"majorTeacher"`
		MajorId                  interface{} `json:"majorId"`
		MajorName                interface{} `json:"majorName"`
		MajorField               interface{} `json:"majorField"`
		Semester                 interface{} `json:"semester"`
		PlanExtra                interface{} `json:"planExtra"`
		MajorTeacherId           interface{} `json:"majorTeacherId"`
		IsSysDefault             int         `json:"isSysDefault"`
		InternshipForm           interface{} `json:"internshipForm"`
		TeacherName              interface{} `json:"teacherName"`
		TeacherId                interface{} `json:"teacherId"`
		Mobile                   interface{} `json:"mobile"`
		IsCopyAllocate           interface{} `json:"isCopyAllocate"`
		IsCopy                   interface{} `json:"isCopy"`
		IsShowUpDel              interface{} `json:"isShowUpDel"`
		IsBuyInsurance           interface{} `json:"isBuyInsurance"`
		StuItemIds               interface{} `json:"stuItemIds"`
		SelfMultiple             interface{} `json:"selfMultiple"`
		SchoolTeacher            interface{} `json:"schoolTeacher"`
		CompanyMultiple          interface{} `json:"companyMultiple"`
		MultipleTheory           interface{} `json:"multipleTheory"`
		PracticeState            int         `json:"practiceState"`
		AboutType                interface{} `json:"aboutType"`
		PracticeTeas             interface{} `json:"practiceTeas"`
		PracticeStateNum         interface{} `json:"practiceStateNum"`
		ProgramId                interface{} `json:"programId"`
		AttachmentsList          interface{} `json:"attachmentsList"`
		IsTalentPlan             interface{} `json:"isTalentPlan"`
		Comment                  interface{} `json:"comment"`
		AuditName                interface{} `json:"auditName"`
		IsApply                  int         `json:"isApply"`
		LevelEntity              interface{} `json:"levelEntity"`
		InsuranceList            interface{} `json:"insuranceList"`
	} `json:"data"`
}