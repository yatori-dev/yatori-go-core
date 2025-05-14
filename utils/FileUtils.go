package utils

import (
	"fmt"
	"image"
	"image/png"
	"os"
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
