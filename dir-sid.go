package main

import (
	"fmt"
	"io/ioutil"
	"os"

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

	i, err := DirSid(path)
	if err != nil {
		panic(err)
	}

	g := map[string]int64{}

	for _, fi := range i {
		fmt.Printf("%v\t%s\t%s\n", fi.Size, fi.OwnerSid, fi.Path)
		g[fi.OwnerSid] += fi.Size
	}

	for sid, size := range g {
		fmt.Printf("\033[32m%d\t%s\n\033[0m", size, sid)
	}

}

func DirSid(path string) ([]FileInfo, error) {
	var fileInfo []FileInfo

	fileStat, err := os.Stat(path)
	if err != nil {
		return nil, err
	}

	if fileStat.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return nil, err
		}
		for _, f := range files {
			i, err := DirSid(path + "\\" + f.Name())
			if err != nil {
				return nil, err
			}
			fileInfo = append(fileInfo, i...)
		}
	} else {
		fileInfo = append(fileInfo, FileInfo{Path: path, OwnerSid: GetFileSid(path), Size: fileStat.Size()})
	}
	return fileInfo, nil
}

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
