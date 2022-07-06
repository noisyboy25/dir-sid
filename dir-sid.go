package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/hectane/go-acl/api"
	"golang.org/x/sys/windows"
)

type FileInfo struct {
	Path     string
	OwnerSid string
	Size     int64
}

func main() {
	args := os.Args
	path := "."

	if len(args) > 1 && args[1] != "" {
		path = args[1]
	}

	start := time.Now()
	i := GetDirInfo(path)
	t := time.Since(start)

	g := map[string]int64{}

	for _, fi := range i {
		g[fi.OwnerSid] += fi.Size
	}

	for sid, size := range g {
		fmt.Printf("\033[32m%d\t%s\n\033[0m", size, sid)
	}

	fmt.Printf("Completed in: %v\n", t)
}

// GetDirInfo returns the directory information for the given path
func GetDirInfo(path string) (fileInfo []FileInfo) {
	wg := &sync.WaitGroup{}
	c := make(chan FileInfo)

	wg.Add(1)
	go func() {
		pushFileInfo(path, wg, c)
		wg.Wait()
		close(c)
	}()

	for i := range c {
		fileInfo = append(fileInfo, i)
	}

	return fileInfo
}

// TODO: error handling
//
// pushFileInfo pushes the file information to the channel
func pushFileInfo(path string, wg *sync.WaitGroup, c chan FileInfo) {
	defer wg.Done()
	fileStat, _ := os.Stat(path)

	if fileStat.IsDir() {
		files, _ := ioutil.ReadDir(path)
		for _, f := range files {
			wg.Add(1)
			go pushFileInfo(path+"\\"+f.Name(), wg, c)
		}
	} else {
		c <- FileInfo{Path: path, OwnerSid: GetFileSid(path), Size: fileStat.Size()}
	}
}

// GetDirInfo returns the file information for the given path
func GetFileSid(path string) string {
	var owner *windows.SID

	err := api.GetNamedSecurityInfo(
		path,
		api.SE_FILE_OBJECT,
		api.OWNER_SECURITY_INFORMATION,
		&owner,
		nil,
		nil,
		nil,
		nil,
	)
	if err != nil {
		fmt.Println(err)
	}

	return owner.String()
}
