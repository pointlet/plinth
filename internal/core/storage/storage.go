// Package storage implements the local filesystem storage engine.
package storage

type StorageServer struct{}

func Start() *StorageServer {
	println("Storage server strating...")
	return &StorageServer{}
}

func Put() {}

func Get() {}

func Stat() {}

func List() {}

func Delete() {}

func Move() {}

func Copy() {}
