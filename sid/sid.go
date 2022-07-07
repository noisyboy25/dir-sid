package sid

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/hectane/go-acl/api"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sys/windows"
)

type Info struct {
	Path     string
	OwnerSid string
	Size     int64
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
