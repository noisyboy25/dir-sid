//go:build windows

package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/noisyboy25/dir-sid/sid"
)

func main() {
	args := os.Args
	path := "."

	if len(args) > 1 && args[1] != "" {
		path = args[1]
	}

	start := time.Now()
	i, err := sid.GetDirInfo(path, context.Background())
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
