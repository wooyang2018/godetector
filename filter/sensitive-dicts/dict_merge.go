//合并从网上搜集的敏感词库
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strings"
)

func merge(dirPath string) map[string]struct{} {
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}
	resSet := make(map[string]struct{})
	for _, file := range files {
		curFile := dirPath + "/" + file.Name()
		readFileLines(curFile, resSet)
		fmt.Printf("curFile: %s, len(resSet): %d\n", curFile, len(resSet))
	}
	WriteMaptoFile(resSet, dirPath+".txt")
	return resSet
}

func mergeAll() {
	var dirPath string = "./其他"
	files, err := os.ReadDir(dirPath)
	if err != nil {
		log.Fatal(err)
	}
	resSet := make(map[string]struct{})
	for _, file := range files {
		curFile := dirPath + "/" + file.Name()
		readFileLines(curFile, resSet)
		fmt.Printf("curFile: %s, len(resSet): %d\n", curFile, len(resSet))
	}
	dirNameList := []string{
		"./色情",
		"./网址",
		"./民生",
		"./暴恐",
		"./政治",
		"./广告",
		"./色情",
	}
	for _, dirName := range dirNameList {
		tempSet := merge(dirName)
		for k := range resSet {
			if _, ok := tempSet[k]; ok {
				fmt.Printf("redundant word in %s: %s\n", dirName, k)
				delete(resSet, k)
			}
		}
	}
	WriteMaptoFile(resSet, dirPath+".txt")
}

func readFileLines(filePath string, resSet map[string]struct{}) error {
	var flag struct{}
	f, err := os.Open(filePath)
	if err != nil {
		log.Fatalf("file %s not found", filePath)
		return err
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		strline := scanner.Text()
		strline = strings.TrimRight(strline, ",")
		resSet[strline] = flag
	}
	return nil
}

func WriteMaptoFile(m map[string]struct{}, filePath string) error {
	f, err := os.Create(filePath)
	if err != nil {
		fmt.Printf("create map file error: %v\n", err)
		return err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	for k := range m {
		// lineStr := fmt.Sprintf("%s^%s", k, v)
		fmt.Fprintln(w, k)
	}
	return w.Flush()
}

func main() {
	// merge("./色情")
	// merge("./网址")
	// merge("./民生")
	// merge("./暴恐")
	// merge("./政治")
	// merge("./广告")
	// merge("./色情")
	// merge("./其他")
	mergeAll()
}
