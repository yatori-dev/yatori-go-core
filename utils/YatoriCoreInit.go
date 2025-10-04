package utils

import (
	"embed"
	_ "embed"
	"log"
	"os"

	"github.com/Changbaiqi/ddddocr-go/utils"
)

// 首次调用必须要先进行初始化
var assets embed.FS

// 数据列表
var assetsList = []string{}

func YatoriCoreInit() {
	//加载必要的文件资源
	loadAssets()
	//加载AI环境
	loadAiEnvironment()
}

// 加载必要的文件资源
func loadAssets() {
	//检查文件是否已经复制到本地
	for _, fileName := range assetsList {
		exists, _ := PathExists("./assets/" + fileName)
		if !exists {
			writeAssetsToDisk() // 确保文件都加载了
			break
		}
	}
}

// 将必要文件复制到当前目录下
func writeAssetsToDisk() {
	PathExistForCreate("./assets")

	for _, fileName := range assetsList {
		resource, err1 := assets.ReadFile("assets/" + fileName)
		if err1 != nil {
			log.Println(err1)
		}
		wf_status := os.WriteFile("./assets/"+fileName, resource, 0644)
		if wf_status != nil {
			log.Fatal(wf_status)
		}
	}
}

// 加载AI环境
func loadAiEnvironment() {
	//初始化AI库
	utils.DDDDOcrCoreInit()
}
