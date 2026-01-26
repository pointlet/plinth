// Package storage implements the local filesystem storage engine.
package storage

import (
	"os"
	"path/filepath"
)

const (
	OwnerCanWriteEveryoneCanReadPermissions   = 0o644
	OwnerCanDoAllEveryoneCanAccessPermissions = 0o755
	OnlyOwnerCanReadAndWritePermissions       = 0o600
	OnyOwnerCanAccessPermissions              = 0o700
)

func WriteRawFile(basePath string, directoryAndFilename string, rawFile []byte) error {
	directories := filepath.Dir(directoryAndFilename)
	err := os.MkdirAll(
		filepath.Join(basePath, directories),
		OwnerCanDoAllEveryoneCanAccessPermissions,
	)
	if err != nil {
		return err
	}

	err = os.WriteFile(
		filepath.Join(basePath, directoryAndFilename),
		rawFile,
		OwnerCanWriteEveryoneCanReadPermissions,
	)
	if err != nil {
		return err
	}

	return nil
}

func BatchWriteRawFiles(bashPath string, directoryAndFilename []string, rawFiles []byte) {
	// implement me
}

func ReadRawFile( /* what args do I need?*/ ) {}

func BatchReadRawFile( /* what args do I need?*/ ) {}

func ListDirectoryContents(basePath string, directoryAndFilename string) ([]os.DirEntry, error) {
	entries, err := os.ReadDir(filepath.Join(basePath, directoryAndFilename))
	if err != nil {
		return nil, err
	}

	return entries, nil
}

func Delete( /* what args do I need?*/ ) {}

func ReadMetadata( /* what args do I need?*/ ) {}

func Exists( /* what args do I need?*/ ) {}

func UpdateMetadata( /* what args do I need?*/ ) {}

func Move( /* what args do I need?*/ ) {}

func Copy( /* what args do I need?*/ ) {}

func BatchDelete( /* what args do I need?*/ ) {}
