package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func FileOrDirExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CreateDirs 创建目标目录及其子目录
func CreateDirs(basePath string) error {
	InfoLog("Creating directories")
	//清空目标文件夹
	if err := os.RemoveAll(basePath); err != nil {
		log.Fatal("Error removing target directory", err.Error())
	}
	if err := os.Mkdir(basePath, os.ModeDir); err != nil {
		log.Fatal("Error when creating target dir:" + err.Error())
	}

	//重新创建所需文件
	years := []int{2021, 2022}
	month, day := 13, 32
	//年
	for i := 0; i < len(years); i++ {
		dirPath := filepath.Join(basePath, fmt.Sprint(years[i]))
		if err := os.Mkdir(dirPath, os.ModePerm); err != nil {
			log.Fatal("Error when creating target dir:" + err.Error())
			return err
		}
		//月
		for j := 1; j < month; j++ {
			dirPath1 := filepath.Join(dirPath, fmt.Sprint(j))
			if err := os.Mkdir(dirPath1, os.ModePerm); err != nil {
				log.Fatal("Error when creating target dir:" + err.Error())
				return err
			}
			//日
			for k := 1; k < day; k++ {
				dirPath2 := filepath.Join(dirPath1, fmt.Sprint(k))
				if err := os.Mkdir(dirPath2, os.ModePerm); err != nil {
					log.Fatal("Error when creating target dir:" + err.Error())
					return err
				}
			}
		}
	}
	return nil
}

// DeleteEmptyDirs 删除目标目录中的空目录及其子目录
func DeleteEmptyDirs(filePath string) error {
	//删除子目录
	dirs, err := os.ReadDir(filePath)
	if err != nil {
		log.Println("Error when reading dir:" + err.Error())
		return err
	}
	for _, d := range dirs {
		if d.IsDir() {
			err = DeleteEmptyDirs(filepath.Join(filePath, d.Name()))
			if err != nil {
				log.Println("Error when removing dir:" + err.Error())
				return err
			}
		}
	}
	//删除本目录
	dirs, err = os.ReadDir(filePath)
	if err != nil {
		log.Println("Error when reading dir:" + err.Error())
		return err
	}
	if len(dirs) == 0 {
		err = os.RemoveAll(filePath)
		if err != nil {
			log.Println("Error when removing dir:" + err.Error())
			return err
		}
		return nil
	}
	return nil
}

func ClearFolder(path string) error {
	fileInfo, err := os.Stat(path)
	if os.IsNotExist(err) {
		err := os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
		return nil
	}

	if fileInfo.IsDir() {
		entries, err := os.ReadDir(path)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			entryPath := filepath.Join(path, entry.Name())
			err := os.RemoveAll(entryPath)
			if err != nil {
				continue
			}
		}
	} else {
		err := os.Remove(path)
		if err != nil {
			return err
		}

		err = os.MkdirAll(path, 0755)
		if err != nil {
			return err
		}
	}
	return nil
}

func ParseTimeToInt(timeStr string) int64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0 // 如果格式不正确，返回0
	}

	hour, err1 := strconv.Atoi(parts[0])
	minute, err2 := strconv.Atoi(parts[1])
	second, err3 := strconv.Atoi(parts[2])

	if err1 != nil || err2 != nil || err3 != nil {
		return 0 // 如果转换失败，返回0
	}

	return int64(hour*60*60 + minute*60 + second)
}

// TimeStringToUnixMillis 将 "hh:mm:ss" 格式的时间字符串转换为毫秒级 Unix 时间戳
// 假定日期为2025年9月1号，时区为UTC
func TimeStringToUnixMillis(timeStr string) int64 {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 3 {
		return 0 // 如果格式不正确，返回0
	}

	hour, err1 := strconv.Atoi(parts[0])
	minute, err2 := strconv.Atoi(parts[1])
	second, err3 := strconv.Atoi(parts[2])

	if err1 != nil || err2 != nil || err3 != nil {
		return 0 // 如果转换失败，返回0
	}

	// 验证时间范围
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 || second < 0 || second > 59 {
		return 0 // 时间范围不合法，返回0
	}

	// 创建2025年9月1号的时间对象（UTC时区）
	baseDate := time.Date(2025, 9, 1, hour, minute, second, 0, time.UTC)
	
	// 返回毫秒级时间戳
	return baseDate.UnixMilli()
}
