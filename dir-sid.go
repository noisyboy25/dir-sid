package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/hectane/go-acl/api"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/windows"
)

type Info struct {
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
	i, err := GetDirInfo(path, context.Background())
	if err != nil {
		panic(err)
	}

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
func GetDirInfo(path string, ctx context.Context) (fileInfo []Info, err error) {
	eg, ctx := errgroup.WithContext(ctx)
	c := make(chan Info)

	go func() {
		defer close(c)
		eg.Go(func() error {
			err := pushFileInfo(path, eg, ctx, c)
			return err
		})
		eg.Wait()
	}()

	for i := range c {
		fileInfo = append(fileInfo, i)
	}

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return fileInfo, err
}

// pushFileInfo pushes the file information to the channel
func pushFileInfo(path string, eg *errgroup.Group, ctx context.Context, c chan Info) error {
	fileStat, err := os.Stat(path)
	if err != nil {
		return err
	}

	if fileStat.IsDir() {
		files, err := ioutil.ReadDir(path)
		if err != nil {
			return err
		}

		for _, f := range files {
			func(f os.FileInfo) {
				eg.Go(func() error {
					err := pushFileInfo(path+"\\"+f.Name(), eg, ctx, c)
					return err
				})
			}(f)
		}
	} else {
		select {
		case c <- Info{Path: path, OwnerSid: GetFileSid(path), Size: fileStat.Size()}:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
	return nil
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
