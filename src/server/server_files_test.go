package server

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/ihexxa/quickshare/src/client"
	q "github.com/ihexxa/quickshare/src/handlers"
	"github.com/ihexxa/quickshare/src/handlers/fileshdr"
)

func TestFileHandlers(t *testing.T) {
	addr := "http://127.0.0.1:8686"
	root := "testData"
	config := `{
		"users": {
			"enableAuth": true,
			"minUserNameLen": 2,
			"minPwdLen": 4,
			"captchaEnabled": false,
			"uploadSpeedLimit": 409600,
			"downloadSpeedLimit": 409600,
			"spaceLimit": 1000,
			"limiterCapacity": 1000,
			"limiterCyc": 1000
		},
		"server": {
			"debug": true
		},
		"fs": {
			"root": "testData"
		}
	}`

	adminName := "qs"
	adminPwd := "quicksh@re"
	os.Setenv("DEFAULTADMIN", adminName)
	os.Setenv("DEFAULTADMINPWD", adminPwd)

	os.RemoveAll(root)
	err := os.MkdirAll(root, 0700)
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	srv := startTestServer(config)
	defer srv.Shutdown()
	fs := srv.depsFS()

	if !waitForReady(addr) {
		t.Fatal("fail to start server")
	}

	usersCl := client.NewSingleUserClient(addr)
	resp, _, errs := usersCl.Login(adminName, adminPwd)
	if len(errs) > 0 {
		t.Fatal(errs)
	} else if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
	token := client.GetCookie(resp.Cookies(), q.TokenCookie)
	cl := client.NewFilesClient(addr, token)

	// TODO: remove all files under home folder before testing
	// or the count of files is incorrect
	t.Run("ListHome", func(t *testing.T) {
		files := map[string]string{
			"qs/files/home_file1": "12345678",
			"qs/files/home_file2": "12345678",
		}

		for filePath, content := range files {
			assertUploadOK(t, filePath, content, addr, token)

			err = fs.Sync()
			if err != nil {
				t.Fatal(err)
			}
		}

		resp, lhResp, errs := cl.ListHome()
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if resp.StatusCode != 200 {
			t.Fatal(resp.StatusCode)
		} else if lhResp.Cwd != "qs/files" {
			t.Fatalf("incorrect ListHome cwd %s", lhResp.Cwd)
		} else if len(lhResp.Metadatas) != len(files) {
			for _, metadata := range lhResp.Metadatas {
				fmt.Printf("%v\n", metadata)
			}
			t.Fatalf("incorrect ListHome content %d", len(lhResp.Metadatas))
		}

		infos := map[string]*fileshdr.MetadataResp{}
		for _, metadata := range lhResp.Metadatas {
			infos[metadata.Name] = metadata
		}

		if infos["home_file1"].Size != int64(len(files["qs/files/home_file1"])) {
			t.Fatalf("incorrect file size %d", infos["home_file1"].Size)
		} else if infos["home_file1"].IsDir {
			t.Fatal("incorrect item type")
		}
		if infos["home_file2"].Size != int64(len(files["qs/files/home_file2"])) {
			t.Fatalf("incorrect file size %d", infos["home_file2"].Size)
		} else if infos["home_file2"].IsDir {
			t.Fatal("incorrect item type")
		}
	})

	t.Run("test uploading files with duplicated names", func(t *testing.T) {
		files := map[string]string{
			"qs/files/dupdir/dup_file1":     "12345678",
			"qs/files/dupdir/dup_file2.ext": "12345678",
		}
		renames := map[string]string{
			"qs/files/dupdir/dup_file1":     "qs/files/dupdir/dup_file1_1",
			"qs/files/dupdir/dup_file2.ext": "qs/files/dupdir/dup_file2_1.ext",
		}

		for filePath, content := range files {
			for i := 0; i < 2; i++ {
				assertUploadOK(t, filePath, content, addr, token)

				err = fs.Sync()
				if err != nil {
					t.Fatal(err)
				}

				if i == 0 {
					assertDownloadOK(t, filePath, content, addr, token)
				} else if i == 1 {
					renamedFilePath, ok := renames[filePath]
					if !ok {
						t.Fatal("new name not found")
					}
					assertDownloadOK(t, renamedFilePath, content, addr, token)
				}
			}
		}
	})

	t.Run("test files APIs: Create-UploadChunk-UploadStatus-Metadata-Delete", func(t *testing.T) {
		for filePath, content := range map[string]string{
			"qs/files/path1/f1.md":       "1111 1111 1111 1111",
			"qs/files/path1/path2/f2.md": "1010 1010 1111 0000 0010",
		} {
			fileSize := int64(len([]byte(content)))
			// create a file
			res, _, errs := cl.Create(filePath, fileSize)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}

			// check uploading file
			uploadFilePath := q.UploadPath(adminName, filePath)
			info, err := fs.Stat(uploadFilePath)
			if err != nil {
				t.Fatal(err)
			} else if info.Name() != filepath.Base(uploadFilePath) {
				t.Fatal(info.Name(), filepath.Base(uploadFilePath))
			}

			// upload a chunk
			i := 0
			contentBytes := []byte(content)
			for i < len(contentBytes) {
				right := i + rand.Intn(3) + 1
				if right > len(contentBytes) {
					right = len(contentBytes)
				}

				chunk := contentBytes[i:right]
				chunkBase64 := base64.StdEncoding.EncodeToString(chunk)
				res, _, errs = cl.UploadChunk(filePath, chunkBase64, int64(i))
				i = right
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if res.StatusCode != 200 {
					t.Fatal(res.StatusCode)
				}

				if int64(i) != fileSize {
					_, statusResp, errs := cl.UploadStatus(filePath)
					if len(errs) > 0 {
						t.Fatal(errs)
					} else if statusResp.Path != filePath ||
						statusResp.IsDir ||
						statusResp.FileSize != fileSize ||
						statusResp.Uploaded != int64(i) {
						t.Fatal("incorrect uploadinfo info", statusResp)
					}
				}
			}

			err = fs.Sync()
			if err != nil {
				t.Fatal(err)
			}

			// check uploaded file
			// fsFilePath := filepath.Join("0", filePath)
			info, err = fs.Stat(filePath)
			if err != nil {
				t.Fatal(err)
			} else if info.Name() != filepath.Base(filePath) {
				t.Fatal(info.Name(), filepath.Base(filePath))
			}

			// metadata
			_, mRes, errs := cl.Metadata(filePath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if mRes.Name != info.Name() ||
				mRes.IsDir != info.IsDir() ||
				mRes.Size != info.Size() {
				// TODO: modTime is not checked
				t.Fatal("incorrect uploaded info", mRes)
			}

			// delete file
			res, _, errs = cl.Delete(filePath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}
	})

	t.Run("test dirs APIs: Mkdir-Create-UploadChunk-List", func(t *testing.T) {
		for dirPath, files := range map[string]map[string]string{
			"qs/files/dir/path1": map[string]string{
				"f1.md": "11111",
				"f2.md": "22222222222",
			},
			"qs/files/dir/path2/path2": map[string]string{
				"f3.md": "3333333",
			},
		} {
			res, _, errs := cl.Mkdir(dirPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}

			for fileName, content := range files {
				filePath := filepath.Join(dirPath, fileName)
				assertUploadOK(t, filePath, content, addr, token)
			}

			err = fs.Sync()
			if err != nil {
				t.Fatal(err)
			}

			_, lResp, errs := cl.List(dirPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			}
			for _, metadata := range lResp.Metadatas {
				content, ok := files[metadata.Name]
				if !ok {
					t.Fatalf("%s not found", metadata.Name)
				} else if int64(len(content)) != metadata.Size {
					t.Fatalf("size not match %d %d \n", len(content), metadata.Size)
				}
			}
		}
	})

	t.Run("test operation APIs: Mkdir-Create-UploadChunk-Move-List", func(t *testing.T) {
		srcDir := "qs/files/move/src"
		dstDir := "qs/files/move/dst"

		for _, dirPath := range []string{srcDir, dstDir} {
			res, _, errs := cl.Mkdir(dirPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}

		files := map[string]string{
			"f1.md": "111",
			"f2.md": "22222",
		}

		for fileName, content := range files {
			oldPath := filepath.Join(srcDir, fileName)
			newPath := filepath.Join(dstDir, fileName)
			// fileSize := int64(len([]byte(content)))
			assertUploadOK(t, oldPath, content, addr, token)

			res, _, errs := cl.Move(oldPath, newPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}

		err = fs.Sync()
		if err != nil {
			t.Fatal(err)
		}

		_, lResp, errs := cl.List(dstDir)
		if len(errs) > 0 {
			t.Fatal(errs)
		}
		for _, metadata := range lResp.Metadatas {
			content, ok := files[metadata.Name]
			if !ok {
				t.Fatalf("%s not found", metadata.Name)
			} else if int64(len(content)) != metadata.Size {
				t.Fatalf("size not match %d %d \n", len(content), metadata.Size)
			}
		}
	})

	t.Run("test download APIs: Download(normal, ranges)", func(t *testing.T) {
		for filePath, content := range map[string]string{
			"qs/files/download/path1/f1":    "123456",
			"qs/files/download/path1/path2": "12345678",
		} {
			assertUploadOK(t, filePath, content, addr, token)

			err = fs.Sync()
			if err != nil {
				t.Fatal(err)
			}

			assertDownloadOK(t, filePath, content, addr, token)
		}
	})

	t.Run("test sharing APIs: Upload-AddSharing-ListSharings-List-Download-DelSharing-ListSharings", func(t *testing.T) {
		files := map[string]string{
			"qs/files/sharing/path1/f1":    "123456",
			"qs/files/sharing/path2/path2": "12345678",
		}

		for filePath, content := range files {
			assertUploadOK(t, filePath, content, addr, token)
			err = fs.Sync()
			if err != nil {
				t.Fatal(err)
			}
		}

		// add sharings
		sharedPaths := map[string]bool{}
		for filePath := range files {
			dirPath := filepath.Dir(filePath)
			sharedPaths[dirPath] = true

			res, _, errs := cl.AddSharing(dirPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}

		// check listSharings
		res, shRes, errs := cl.ListSharings()
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if res.StatusCode != 200 {
			t.Fatal(res.StatusCode)
		}
		for _, dirPath := range shRes.SharingDirs {
			if !sharedPaths[dirPath] {
				t.Fatalf("sharing %s not found", dirPath)
			}
		}

		for dirPath := range sharedPaths {
			res, _, errs := cl.IsSharing(dirPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}

			res, _, errs = cl.IsSharing(fmt.Sprintf("%s/", dirPath))
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 404 {
				t.Fatal(res.StatusCode)
			}
		}

		for _, dirPath := range shRes.SharingDirs {
			res, lsResp, errs := cl.List(dirPath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
			if lsResp.Cwd != dirPath {
				t.Fatalf("list sharing folder: incorrect cwd (%s)", lsResp.Cwd)
			}
			if len(lsResp.Metadatas) != 1 {
				t.Fatalf("list sharing folder: incorrect metadata size (%d)", len(lsResp.Metadatas))
			}
		}

		for filePath, content := range files {
			assertDownloadOK(t, filePath, content, addr, token)
		}

		i := 0
		for filePath := range files {
			if i++; i%2 == 0 {
				dirPath := filepath.Dir(filePath)
				delete(sharedPaths, dirPath)

				res, _, errs := cl.DelSharing(dirPath)
				if len(errs) > 0 {
					t.Fatal(errs)
				} else if res.StatusCode != 200 {
					t.Fatal(res.StatusCode)
				}
			}
		}

		// check listSharings
		res, shRes, errs = cl.ListSharings()
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if res.StatusCode != 200 {
			t.Fatal(res.StatusCode)
		}
		for _, dirPath := range shRes.SharingDirs {
			if !sharedPaths[dirPath] {
				t.Fatalf("sharing %s not found", dirPath)
			}
		}
	})

	t.Run("test concurrently uploading & downloading", func(t *testing.T) {
		type mockFile struct {
			FilePath string
			Content  string
		}
		wg := &sync.WaitGroup{}

		startClient := func(files []*mockFile) {
			for i := 0; i < 5; i++ {
				for _, file := range files {
					if !assertUploadOK(t, fmt.Sprintf("%s_%d", file.FilePath, i), file.Content, addr, token) {
						break
					}

					err = fs.Sync()
					if err != nil {
						t.Fatal(err)
					}

					if !assertDownloadOK(t, fmt.Sprintf("%s_%d", file.FilePath, i), file.Content, addr, token) {
						break
					}
				}
			}

			wg.Done()
		}

		for _, clientFiles := range [][]*mockFile{
			[]*mockFile{
				&mockFile{"concurrent/p0/f0", "00"},
				&mockFile{"concurrent/f0.md", "0000 0000 0000 0"},
			},
			[]*mockFile{
				&mockFile{"concurrent/p1/f1", "11"},
				&mockFile{"concurrent/f1.md", "1111 1111 1"},
			},
			[]*mockFile{
				&mockFile{"concurrent/p2/f2", "22"},
				&mockFile{"concurrent/f2.md", "222"},
			},
		} {
			wg.Add(1)
			go startClient(clientFiles)
		}

		wg.Wait()
	})

	t.Run("test uploading APIs: ListUploadings, Create, ListUploadings, DelUploading", func(t *testing.T) {
		// it should return no error even no file is uploaded
		res, lResp, errs := cl.ListUploadings()
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if res.StatusCode != 200 {
			t.Fatal(res.StatusCode)
		}

		files := map[string]string{
			"qs/files/uploadings/path1/f1":    "123456",
			"qs/files/uploadings/path1/path2": "12345678",
		}

		for filePath, content := range files {
			fileSize := int64(len([]byte(content)))
			res, _, errs := cl.Create(filePath, fileSize)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}

		res, lResp, errs = cl.ListUploadings()
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if res.StatusCode != 200 {
			t.Fatal(res.StatusCode)
		}

		gotInfos := map[string]*fileshdr.UploadInfo{}
		for _, info := range lResp.UploadInfos {
			gotInfos[info.RealFilePath] = info
		}
		for filePath, content := range files {
			info, ok := gotInfos[filePath]
			if !ok {
				t.Fatalf("uploading(%s) not found", filePath)
			} else if info.Uploaded != 0 {
				t.Fatalf("uploading(%s) uploaded is not correct", filePath)
			} else if info.Size != int64(len([]byte(content))) {
				t.Fatalf("uploading(%s) size is not correct", filePath)
			}
		}

		for filePath := range files {
			res, _, errs := cl.DelUploading(filePath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}

		res, lResp, errs = cl.ListUploadings()
		if len(errs) > 0 {
			t.Fatal(errs)
		} else if res.StatusCode != 200 {
			t.Fatal(res.StatusCode)
		} else if len(lResp.UploadInfos) != 0 {
			t.Fatalf("info is not deleted, info len(%d)", len(lResp.UploadInfos))
		}
	})

	t.Run("test uploading APIs: Create, Stop, UploadChunk", func(t *testing.T) {
		// cl := client.NewFilesClient(addr)

		files := map[string]string{
			"qs/files/uploadings/path1/f1": "12345678",
		}

		for filePath, content := range files {
			fileSize := int64(len([]byte(content)))
			res, _, errs := cl.Create(filePath, fileSize)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}

			chunks := [][]byte{
				[]byte(content)[:fileSize/2],
				[]byte(content)[fileSize/2:],
			}
			offset := int64(0)
			for _, chunk := range chunks {
				base64Content := base64.StdEncoding.EncodeToString(chunk)
				res, _, errs = cl.UploadChunk(filePath, base64Content, offset)
				offset += int64(len(chunk))

				if len(errs) > 0 {
					t.Fatal(errs)
				} else if res.StatusCode != 200 {
					t.Fatal(res.StatusCode)
				}

				err = fs.Close()
				if err != nil {
					t.Fatal(err)
				}
			}

			err = fs.Sync()
			if err != nil {
				t.Fatal(err)
			}

			// metadata
			_, mRes, errs := cl.Metadata(filePath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if mRes.Size != fileSize {
				t.Fatal("incorrect uploaded size", mRes)
			}

			assertDownloadOK(t, filePath, content, addr, token)
		}
	})

	t.Run("test uploading APIs: Create and UploadChunk randomly", func(t *testing.T) {
		// cl := client.NewFilesClient(addr)

		files := map[string]string{
			"qs/files/uploadings/random/path1/f1": "12345678",
			"qs/files/uploadings/random/path1/f2": "87654321",
			"qs/files/uploadings/random/path1/f3": "17654321",
		}

		for filePath, content := range files {
			fileSize := int64(len([]byte(content)))
			res, _, errs := cl.Create(filePath, fileSize)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if res.StatusCode != 200 {
				t.Fatal(res.StatusCode)
			}
		}

		for filePath, content := range files {
			fileSize := int64(len([]byte(content)))

			chunks := [][]byte{
				[]byte(content)[:fileSize/2],
				[]byte(content)[fileSize/2:],
			}
			offset := int64(0)
			for _, chunk := range chunks {
				base64Content := base64.StdEncoding.EncodeToString(chunk)
				res, _, errs := cl.UploadChunk(filePath, base64Content, offset)
				offset += int64(len(chunk))

				if len(errs) > 0 {
					t.Fatal(errs)
				} else if res.StatusCode != 200 {
					t.Fatal(res.StatusCode)
				}

				err = fs.Close()
				if err != nil {
					t.Fatal(err)
				}
			}

			err = fs.Sync()
			if err != nil {
				t.Fatal(err)
			}

			// metadata
			_, mRes, errs := cl.Metadata(filePath)
			if len(errs) > 0 {
				t.Fatal(errs)
			} else if mRes.Size != fileSize {
				t.Fatal("incorrect uploaded size", mRes)
			}

			isEqual, err := compareFileContent(fs, "0", filePath, content)
			if err != nil {
				t.Fatalf("err comparing content: %s", err)
			} else if !isEqual {
				t.Fatalf("file content not equal: %s", filePath)
			}

			assertDownloadOK(t, filePath, content, addr, token)
		}
	})

	resp, _, errs = usersCl.Logout(token)
	if len(errs) > 0 {
		t.Fatal(errs)
	} else if resp.StatusCode != 200 {
		t.Fatal(resp.StatusCode)
	}
}
