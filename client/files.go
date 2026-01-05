package main

import (
	"io/ioutil"
	"os"

	"github.com/rustamniraula90/gop2p/protocol"
)

func ScanShareDir() ([]protocol.FileInfo, error) {
	var files []protocol.FileInfo
	root := "data/share"

	entries, err := ioutil.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return []protocol.FileInfo{}, nil
		}
		return nil, err
	}

	for _, entry := range entries {
		t := "file"
		if entry.IsDir() {
			t = "dir"
		}
		files = append(files, protocol.FileInfo{
			Name: entry.Name(),
			Size: entry.Size(),
			Type: t,
		})
	}
	return files, nil
}
