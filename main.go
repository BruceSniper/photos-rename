/**
 * @author Bruce Zhang
 * @description 批量重命名图片文件工具
 * @date 10:26 2025/11/24
 * @dir
 * @file main.go
 **/

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// 获取指定目录下的所有图片文件
func getImageFiles(dir string) ([]string, error) {
	var imageFiles []string
	imageExts := map[string]bool{
		".nef":  true,
		".dng":  true,
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if imageExts[ext] {
			imageFiles = append(imageFiles, entry.Name())
		}
	}

	// 按文件名排序
	sort.Strings(imageFiles)
	return imageFiles, nil
}

// 重命名图片文件，startNum 为起始编号，digits 为文件名数字位数
func renameImages(dir string, dryRun bool, startNum int, digits int) error {
	// 1. 获取所有图片文件
	imageFiles, err := getImageFiles(dir)
	if err != nil {
		return err
	}

	if len(imageFiles) == 0 {
		return fmt.Errorf("目录中没有找到图片文件")
	}

	fmt.Printf("找到 %d 个图片文件\n\n", len(imageFiles))

	// 2. 生成重命名计划
	var renamePlan []struct {
		oldName string
		newName string
	}

	for i, oldFile := range imageFiles {
		ext := filepath.Ext(oldFile)
		newName := fmt.Sprintf("%0*d%s", digits, startNum+i, ext)

		// 如果新旧文件名相同，跳过
		if oldFile == newName {
			continue
		}

		renamePlan = append(renamePlan, struct {
			oldName string
			newName string
		}{
			oldName: oldFile,
			newName: newName,
		})
	}

	if len(renamePlan) == 0 {
		fmt.Println("所有文件名已经是正确的格式，无需重命名")
		return nil
	}

	// 3. 显示重命名计划
	fmt.Println("重命名计划:")
	fmt.Println("----------------------------------------")
	for _, plan := range renamePlan {
		fmt.Printf("%s -> %s\n", plan.oldName, plan.newName)
	}
	fmt.Println("----------------------------------------")
	fmt.Printf("总共需要重命名 %d 个文件\n\n", len(renamePlan))

	// 4. 如果是预览模式，直接返回
	if dryRun {
		fmt.Println("这是预览模式，没有实际执行重命名操作")
		return nil
	}

	// 5. 执行重命名（使用临时文件名避免冲突）
	fmt.Println("开始重命名...")

	// 第一步：先全部重命名为临时文件名
	tempFiles := make(map[string]string) // 临时文件名 -> 最终文件名
	for i, plan := range renamePlan {
		oldPath := filepath.Join(dir, plan.oldName)
		tempName := fmt.Sprintf("temp_%d_%s", i, plan.newName)
		tempPath := filepath.Join(dir, tempName)

		if err := os.Rename(oldPath, tempPath); err != nil {
			return fmt.Errorf("重命名失败 %s -> %s: %v", plan.oldName, tempName, err)
		}
		tempFiles[tempName] = plan.newName
	}

	// 第二步：将临时文件名改为最终文件名
	for tempName, finalName := range tempFiles {
		tempPath := filepath.Join(dir, tempName)
		finalPath := filepath.Join(dir, finalName)

		if err := os.Rename(tempPath, finalPath); err != nil {
			return fmt.Errorf("重命名失败 %s -> %s: %v", tempName, finalName, err)
		}
	}

	fmt.Println("重命名完成!")
	return nil
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("使用方法:")
		fmt.Println("  go run main.go <目录路径> [起始编号] [digits=位数] [preview]")
		fmt.Println()
		fmt.Println("参数说明:")
		fmt.Println("  目录路径      必填，图片所在目录")
		fmt.Println("  起始编号      可选，文件名起始编号，默认为 1")
		fmt.Println("  digits=N     可选，文件名数字位数，默认为 4")
		fmt.Println("  preview      可选，预览模式，不实际执行重命名")
		fmt.Println()
		fmt.Println("示例:")
		fmt.Println("  go run main.go ./photos                        # 0001, 0002, ...")
		fmt.Println("  go run main.go ./photos 2001                   # 2001, 2002, ...")
		fmt.Println("  go run main.go ./photos digits=6               # 000001, 000002, ...")
		fmt.Println("  go run main.go ./photos 2001 digits=6          # 002001, 002002, ...")
		fmt.Println("  go run main.go ./photos 2001 digits=6 preview  # 预览模式")
		os.Exit(1)
	}

	dir := os.Args[1]
	startNum := 1
	digits := 4
	dryRun := false

	// 解析剩余参数
	for _, arg := range os.Args[2:] {
		if arg == "preview" {
			dryRun = true
		} else if strings.HasPrefix(arg, "digits=") {
			num, err := strconv.Atoi(strings.TrimPrefix(arg, "digits="))
			if err != nil || num < 1 {
				fmt.Printf("错误: 位数 '%s' 无效，必须是大于 0 的整数\n", arg)
				os.Exit(1)
			}
			digits = num
		} else {
			num, err := strconv.Atoi(arg)
			if err != nil || num < 1 {
				fmt.Printf("错误: 起始编号 '%s' 无效，必须是大于 0 的整数\n", arg)
				os.Exit(1)
			}
			startNum = num
		}
	}

	// 检查目录是否存在
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		fmt.Printf("错误: %s 不是有效的目录\n", dir)
		os.Exit(1)
	}

	if dryRun {
		fmt.Println("=== 预览模式 ===")
	} else {
		fmt.Println("=== 执行模式 ===")
	}
	fmt.Printf("起始编号: %0*d, 位数: %d\n\n", digits, startNum, digits)

	if err := renameImages(dir, dryRun, startNum, digits); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}
