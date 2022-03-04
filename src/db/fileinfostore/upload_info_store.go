package fileinfostore

import (
	"encoding/json"
	"errors"

	"github.com/ihexxa/quickshare/src/db"
)

var (
	ErrGreaterThanSize = errors.New("uploaded is greater than file size")
	ErrUploadNotFound  = errors.New("upload info not found")
)

func (fi *FileInfoStore) AddUploadInfo(user, filePath, tmpPath string, fileSize int64) error {
	ns := db.UploadNS(user)
	err := fi.store.AddNamespace(ns)
	if err != nil {
		return err
	}

	_, ok := fi.store.GetStringIn(ns, tmpPath)
	if ok {
		return db.ErrCreateExisting
	}

	info := &db.UploadInfo{
		RealFilePath: filePath,
		Size:         fileSize,
		Uploaded:     0,
	}
	infoBytes, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return fi.store.SetStringIn(ns, tmpPath, string(infoBytes))
}

func (fi *FileInfoStore) SetUploadInfo(user, filePath string, newUploaded int64) error {
	realFilePath, fileSize, _, err := fi.GetUploadInfo(user, filePath)
	if err != nil {
		return err
	} else if newUploaded > fileSize {
		return ErrGreaterThanSize
	}

	newInfo := &db.UploadInfo{
		RealFilePath: realFilePath,
		Size:         fileSize,
		Uploaded:     newUploaded,
	}
	newInfoBytes, err := json.Marshal(newInfo)
	if err != nil {
		return err
	}
	return fi.store.SetStringIn(db.UploadNS(user), filePath, string(newInfoBytes))
}

func (fi *FileInfoStore) GetUploadInfo(user, filePath string) (string, int64, int64, error) {
	ns := db.UploadNS(user)
	infoBytes, ok := fi.store.GetStringIn(ns, filePath)
	if !ok {
		return "", 0, 0, ErrUploadNotFound
	}

	info := &db.UploadInfo{}
	err := json.Unmarshal([]byte(infoBytes), info)
	if err != nil {
		return "", 0, 0, err
	}

	return info.RealFilePath, info.Size, info.Uploaded, nil
}

func (fi *FileInfoStore) DelUploadInfo(user, filePath string) error {
	return fi.store.DelInt64In(db.UploadNS(user), filePath)
}

func (fi *FileInfoStore) ListUploadInfo(user string) ([]*db.UploadInfo, error) {
	ns := db.UploadNS(user)
	if !fi.store.HasNamespace(ns) {
		return nil, nil
	}

	infoMap, err := fi.store.ListStringsIn(ns)
	if err != nil {
		return nil, err
	}

	infos := []*db.UploadInfo{}
	for _, infoStr := range infoMap {
		info := &db.UploadInfo{}
		err = json.Unmarshal([]byte(infoStr), info)
		if err != nil {
			return nil, err
		}
		infos = append(infos, info)
	}

	return infos, nil
}
