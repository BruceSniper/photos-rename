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
	"strings"
)

// 获取指定目录下的所有图片文件
func getImageFiles(dir string) ([]string, error) {
	var imageFiles []string
	imageExts := map[string]bool{
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

// 重命名图片文件
func renameImages(dir string, dryRun bool) error {
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
		newName := fmt.Sprintf("%04d%s", i+1, ext)

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
		fmt.Println("  预览模式: go run main.go <目录路径> preview")
		fmt.Println("  执行重命名: go run main.go <目录路径>")
		fmt.Println("\n示例:")
		fmt.Println("  go run main.go ./photos preview")
		fmt.Println("  go run main.go ./photos")
		os.Exit(1)
	}

	dir := os.Args[1]
	dryRun := len(os.Args) > 2 && os.Args[2] == "preview"

	// 检查目录是否存在
	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		fmt.Printf("错误: %s 不是有效的目录\n", dir)
		os.Exit(1)
	}

	if dryRun {
		fmt.Println("=== 预览模式 ===\n")
	} else {
		fmt.Println("=== 执行模式 ===\n")
	}

	if err := renameImages(dir, dryRun); err != nil {
		fmt.Printf("错误: %v\n", err)
		os.Exit(1)
	}
}
