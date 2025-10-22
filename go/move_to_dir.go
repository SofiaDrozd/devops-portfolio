package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	srcDir := flag.String("source", ".", "Source directory to search")
	dstDir := flag.String("dest", "./organized", "Destination directory")
	noRecursive := flag.Bool("no-recursive", false, "Do not search in subdirectories")
	ext := flag.String("ext", ".txt", "File extension to search for")
	flag.Parse()

	if err := os.MkdirAll(*dstDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	err := filepath.WalkDir(*srcDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			log.Println("Error accessing:", path, err)
			return nil
		}

		// 1. If we are inside the destination directory – skip traversing it completely
		if d.IsDir() && filepath.Clean(path) == filepath.Clean(*dstDir) {
			return filepath.SkipDir
		}

		// 2. If non-recursive mode is enabled – skip all subdirectories except the source root
		if *noRecursive && d.IsDir() && path != *srcDir {
			return filepath.SkipDir
		}

		// 3. If the current file is already inside the destination directory – ignore it
		if rel, _ := filepath.Rel(*dstDir, path); !strings.HasPrefix(rel, "..") {
			return nil
		}

		// 4. For any other directory – allow WalkDir to continue without doing anything special
		if d.IsDir() {
			return nil
		}

		if strings.EqualFold(filepath.Ext(d.Name()), *ext) {
			if err := moveFileWithUniqueName(path, *dstDir); err != nil {
				log.Println("Failed to move file:", path, err)
			}
		}

		return nil
	})

	if err != nil {
		log.Fatal(err)
	}
}

func moveFileWithUniqueName(srcPath, dstDir string) error {
	baseName := filepath.Base(srcPath)
	dstPath := filepath.Join(dstDir, baseName)

	for i := 1; ; i++ {
		if _, err := os.Stat(dstPath); os.IsNotExist(err) {
			break
		}
		dstPath = filepath.Join(dstDir, fmt.Sprintf("%d_%d_%s", time.Now().UnixNano(), i, baseName))
	}

	if err := os.Rename(srcPath, dstPath); err != nil {
		return err
	}

	fmt.Println("Moved:", srcPath, "→", dstPath)
	return nil
}
