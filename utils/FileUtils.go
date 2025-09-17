package utils

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
)

// 检测文件夹或文件是否存在
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// 检测目录是否存在，不存在就创建
func PathExistForCreate(path string) {
	exists, _ := PathExists(path)
	if !exists {
		os.MkdirAll(path, os.ModePerm)
	}
}

// 从文件读取imgage
func ReadImg(imgFile string) (image.Image, error) {
	f, err := os.Open(imgFile)
	if err != nil {
		return nil, err
	}
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}
	f.Close()
	return img, nil
}

// 检测图片是否损坏,损坏为true，没损坏为false
func IsBadImg(imgFile string) bool {
	f, err := os.Open(imgFile)
	defer f.Close()
	if err != nil {
		return true
	}
	_, err1 := png.Decode(f)
	if err1 != nil {
		return true
	}
	return false
}

func DeleteFile(path string) {

	// 删除文件
	err := os.Remove(path)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}

// SaveTextToFile 将 content 写入到 path 指定的文件。
// 如果 appendMode 为 true，则以追加模式写入（文件不存在会创建）；
// 否则以覆盖模式写入（会创建或截断已有文件）。
// fileMode 指定创建文件时的权限（如 0644）。
func SaveTextToFile(path, content string, appendMode bool, fileMode os.FileMode) error {
	// 确保目录存在
	dir := filepath.Dir(path)
	if dir != "." {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("创建目录失败: %w", err)
		}
	}

	if appendMode {
		// 以追加模式打开（不存在则创建）
		f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, fileMode)
		if err != nil {
			return fmt.Errorf("打开文件失败: %w", err)
		}
		defer f.Close()

		if _, err := f.WriteString(content); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
	} else {
		// 覆盖写入（会创建或截断已有文件）
		if err := os.WriteFile(path, []byte(content), fileMode); err != nil {
			return fmt.Errorf("写入文件失败: %w", err)
		}
	}

	return nil
}
