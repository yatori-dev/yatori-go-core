package utils

import (
	"embed"
	_ "embed"
	"io/fs"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"

	ddddocr "github.com/Changbaiqi/ddddocr-go/utils"
)

// 首次调用必须要先进行初始化
//
//go:embed assets/tencentCollect.exe
//go:embed assets/tencentPowSolve.exe
//go:embed assets/node_modules/jsdom/**/*
//go:embed assets/cmap.json
//go:embed assets/glyfHashed.json
var assets embed.FS

// 数据列表
var assetsList = []string{
	"tencentCollect.exe",
	"tencentPowSolve.exe",
	"cmap.json",
	"glyfHashed.json",
}

var coreInitOnce sync.Once

func YatoriCoreInit() {
	coreInitOnce.Do(func() {
		//加载必要的文件资源
		loadAssets()
		//加载AI环境
		loadAiEnvironment()
		//加载必要nodejs模块
		loadNodeModules()
		//创建faces文件夹
		PathExistForCreate("./assets/faces")
	})
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
	ddddocr.DDDDOcrCoreInit()
}

// 加载nodejs必要模块
func loadNodeModules() {
	targetRoot := "assets/node_modules"

	// 遍历 embed.FS 中的 node_modules 文件
	err := fs.WalkDir(assets, "assets/node_modules", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// 计算相对路径
		relPath, err := filepath.Rel("assets/node_modules", path)
		if err != nil {
			return err
		}

		targetPath := filepath.Join(targetRoot, relPath)

		if d.IsDir() {
			// 创建目录
			return os.MkdirAll(targetPath, os.ModePerm)
		} else {
			// 写文件
			data, err := assets.ReadFile(path)
			if err != nil {
				return err
			}
			return ioutil.WriteFile(targetPath, data, 0644)
		}
	})

	if err != nil {
		log.Fatalf("加载 node_modules 失败: %v", err)
	}

	//log.Println("node_modules 已成功加载到 assets 文件夹")
}

// ExtractAssets writes the named embedded legacy assets to dir.
//
// Library callers can use this to materialize only the helper binaries a
// platform needs instead of calling YatoriCoreInit, which extracts the
// complete legacy asset bundle into the process working directory.
func ExtractAssets(dir string, names ...string) error {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	for _, name := range names {
		data, err := assets.ReadFile("assets/" + name)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(dir, name), data, 0644); err != nil {
			return err
		}
	}
	return nil
}
