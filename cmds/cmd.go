package main

import (
	"EVdata/common"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// 循环调用 mapmatching.exe
func runMapMatching(times int, interval time.Duration) error {
	// 获取当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前目录失败: %w", err)
	}

	// 构建 mapmatching.exe 的完整路径
	exePath := filepath.Join(currentDir, "mapmatching.exe")

	// 检查文件是否存在
	if _, err := os.Stat(exePath); os.IsNotExist(err) {
		return fmt.Errorf("mapmatching.exe 不存在于当前目录: %s", currentDir)
	}
	argString := []string{"-d=40", " -wc=500", "-wf=1", "-gcf=200"}

	// 循环执行指定次数
	for i := 0; i <= common.VehicleCount/common.BATCH_SIZE; i++ {
		common.InfoLog("第 %d 次执行 mapmatching.exe...\n", i)

		// 创建命令
		argString = append(argString, "-batch="+strconv.Itoa(i))
		cmd := exec.Command(exePath, argString...)

		// 将标准输出和错误输出重定向到当前进程
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		start := time.Now()
		if err := cmd.Run(); err != nil {
			common.ErrorLog("执行失败: %v\n", err)
		} else {
			common.InfoLog("time elapsed:", time.Since(start))
		}
		time.Sleep(interval)
	}

	return nil
}

func main() {
	err := runMapMatching(10, 5*time.Second)
	if err != nil {
		common.ErrorLog("错误: %v\n", err)
		os.Exit(1)
	}
}
