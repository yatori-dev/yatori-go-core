package ctype

type CardType string

const (
	Video        CardType = "insertvideo"  // 视屏
	Work         CardType = "work"         //作业
	Insertdoc    CardType = "insertdoc"    //内置文档
	Document     CardType = "document"     //文档
	InsertBook   CardType = "insertbook"   //内置书
	Hyperlink    CardType = "hyperlink"    //外部链接
	Insertlive   CardType = "insertlive"   //直播
	Insertbbs    CardType = "insertbbs"    //讨论
	InsertReadV2 CardType = "insertreadV2" //阅读任务点
)
