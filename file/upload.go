package file

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type Upload struct {
	fileHeader  *multipart.FileHeader
	basePath    string
	pathHasBase bool
	domain      string
	File        *File
	Error       error
}

func NewUpload(file *multipart.FileHeader) *Upload {
	now := time.Now()
	u := &Upload{
		fileHeader: file,
		File: &File{
			FullName: file.Filename,
			Size:     file.Size,
			IsDir:    false,
			ModTime:  now.Format("2006-01-02 15:04:05"),
			Mime:     file.Header.Get("Content-Type"),
		},
	}
	extension := filepath.Ext(file.Filename)
	u.File.Name = strings.TrimSuffix(file.Filename, extension)
	u.File.Extension = strings.TrimPrefix(extension, ".")
	u.File.FileType, _, _ = strings.Cut(u.File.Mime, "/")
	if u.File.FileType != "image" && u.File.FileType != "video" {
		u.File.FileType = "other"
	}
	u.File.Type = u.File.FileType
	u.File.Path = fmt.Sprintf("upload/%s/%s/%d%s", u.File.FileType, now.Format("2006/01/02"), now.UnixNano(), extension)
	return u
}

func (u *Upload) SetDir(dir string) *Upload {
	if u.Error != nil || dir == "" {
		return u
	}

	if strings.Contains(dir, "..") {
		u.Error = errors.New("路径不能包含“..”")
		return u
	}

	u.File.Path = fmt.Sprintf("%s/%d", dir, time.Now().UnixNano())
	if u.File.Extension != "" {
		u.File.Path += "." + u.File.Extension
	}

	return u
}

func (u *Upload) SetBaseDir(basePath string, pathHasBase bool) *Upload {
	if u.Error != nil || basePath == "" {
		return u
	}

	u.basePath = basePath
	u.pathHasBase = pathHasBase
	if pathHasBase {
		u.File.Path = filepath.Join(basePath, u.File.Path)
	}
	return u
}

func (u *Upload) SetUrl(domain string) *Upload {
	if u.Error != nil || domain == "" {
		return u
	}
	u.File.Url = domain + "/" + u.File.Path
	return u
}

func (u *Upload) Save() *Upload {
	if u.Error != nil {
		return u
	}

	if u.domain != "" {
		u.File.Url = u.domain + "/" + u.File.Path
	}
	dst := u.File.Path
	if !u.pathHasBase {
		dst = filepath.Join(u.basePath, dst)
	}

	// 创建文件目录
	if err := os.MkdirAll(filepath.Dir(strings.TrimPrefix(dst, "/")), os.ModePerm); err != nil {
		u.Error = err
		return u
	}
	// u.Error = c.SaveUploadedFile(u.fileHeader, dst)

	fileReader, err := u.fileHeader.Open()
	if err != nil {
		u.Error = err
		return u
	}
	defer fileReader.Close()

	// 创建目标文件
	destFile, err := os.Create(dst)
	if err != nil {
		u.Error = err
		return u
	}
	defer destFile.Close()

	if _, err = fileReader.Seek(0, io.SeekStart); err != nil {
		// 如果Seek失败，重新打开文件
		fileReader.Close()
		fileReader, err = u.fileHeader.Open()
		if err != nil {
			u.Error = err
			return u
		}
		defer fileReader.Close()
	}

	// 复制文件内容
	_, err = io.Copy(destFile, fileReader)
	if err != nil {
		u.Error = err
	}

	if !strings.HasPrefix(u.File.Path, "/") {
		u.File.Path = "/" + u.File.Path
	}

	return u
}

func (u *Upload) Limit(maxImageSize int64, maxVideoSize int64, maxOtherSize int64, allowExtensions []string) *Upload {
	if u.File.Size == 0 {
		u.Error = errors.New("不能上传空文件")
		return u
	}

	if len(u.File.FullName) > 255 {
		u.Error = errors.New("文件名不能超过255个字符")
		return u
	}

	switch u.File.FileType {
	case "image":
		if u.File.Size > maxImageSize*1024*1024 {
			u.Error = fmt.Errorf("图片不能超过%dM", maxImageSize)
			return u
		}
	case "video":
		if u.File.Size > maxVideoSize*1024*1024 {
			u.Error = fmt.Errorf("视频不能超过%dM", maxVideoSize)
			return u
		}
	case "other":
		if u.File.Size > maxOtherSize*1024*1024 {
			u.Error = fmt.Errorf("文件不能超过%dM", maxOtherSize)
			return u
		}
	}

	if !slices.Contains(allowExtensions, u.File.Extension) {
		u.Error = fmt.Errorf("%s格式不允许上传", u.File.Extension)
		return u
	}

	return u
}

func (u *Upload) GetFile() (*File, error) {
	return u.File, u.Error
}
