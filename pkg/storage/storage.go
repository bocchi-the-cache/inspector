package storage

import (
	"strings"

	diskv "github.com/peterbourgon/diskv/v3"
	"github.com/spf13/viper"
)

var DefaultDiskStorage *DiskStorage

func Init() {
	// Disk K-V storage storage
	storage := diskv.New(diskv.Options{
		BasePath:          viper.GetString("storage.base_case_path"),
		AdvancedTransform: DiskCacheAdvancedTransform,
		InverseTransform:  DiskCacheInverseTransform,
		CacheSizeMax:      1024 * 1024 * 1024 * 1, // 1GB
	})
	DefaultDiskStorage = &DiskStorage{
		C: storage,
	}
}

type DiskStorage struct {
	C *diskv.Diskv
}

func Write(key string, content []byte) error {
	return DefaultDiskStorage.Write(key, content)
}

func (d *DiskStorage) Write(key string, content []byte) error {
	return d.C.Write(key, content)
}

func Read(key string) ([]byte, bool) {
	return DefaultDiskStorage.Read(key)
}

func (d *DiskStorage) Read(key string) ([]byte, bool) {
	content, err := d.C.Read(key)
	if err != nil {
		return nil, false
	}
	return content, true
}

func DiskCacheAdvancedTransform(key string) *diskv.PathKey {
	path := strings.Split(key, "/")
	last := len(path) - 1
	return &diskv.PathKey{
		Path:     path[:last],
		FileName: path[last] + ".file",
	}
}

func DiskCacheInverseTransform(pathKey *diskv.PathKey) (key string) {
	txt := pathKey.FileName[len(pathKey.FileName)-5:]
	if txt != ".file" {
		panic("Invalid file found in storage folder!")
	}
	return strings.Join(pathKey.Path, "/") + pathKey.FileName[:len(pathKey.FileName)-4]
}
