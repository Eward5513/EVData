package common

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

func FileOrDirExists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// CreateDirs 创建目标目录及其子目录
func CreateDirs(basePath string) error {
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
		//log.Println(basePath)
		//log.Println(years[i])
		dirPath := filepath.Join(basePath, fmt.Sprint(years[i]))
		//log.Println(dirPath)
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
