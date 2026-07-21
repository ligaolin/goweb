package file

import (
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
)

const UPLOAD_DIR = "upload"

type Config struct {
	MaxImageSize int64    // 最大图片大小
	MaxVideoSize int64    // 最大视频大小
	MaxOtherSize int64    // 最大其他文件大小
	Extensions   []string // 允许的文件扩展名
	BasePath     string   // 基础路径
	PathHasBase  bool     // 访问路径是否包含基础路径
	Domain       string   // 域名
}

type Upload struct {
	fileHeader *multipart.FileHeader
	config     *Config
	Dir        string
	File       *File
	Error      error
}

func NewUpload(file *multipart.FileHeader, config *Config) *Upload {
	now := time.Now()
	config.BasePath = strings.TrimPrefix(config.BasePath, "\\/")
	u := &Upload{
		config:     config,
		fileHeader: file,
		File: &File{
			FullName: file.Filename,
			Size:     file.Size,
			IsDir:    false,
			ModTime:  now.Format("2006-01-02 15:04:05"),
			Mime:     file.Header.Get("Content-Type"),
		},
	}
	if err := u.detectMimeType(); err != nil {
		u.Error = err
		return u
	}
	extension := filepath.Ext(file.Filename)
	u.File.Name = strings.TrimSuffix(file.Filename, extension)
	u.File.Extension = strings.TrimPrefix(extension, ".")
	u.File.FileType, _, _ = strings.Cut(u.File.Mime, "/")
	if u.File.FileType != "image" && u.File.FileType != "video" {
		u.File.FileType = "other"
	}
	u.File.Type = u.File.FileType
	u.Dir = fmt.Sprintf("%s/%s/%s", UPLOAD_DIR, u.File.FileType, now.Format("2006/01/02"))
	return u
}

// Update 更新文件到指定路径：先保存新文件到临时路径，再删除旧文件，最后移动到目标路径
// targetPath 为目标文件路径（相对于 BasePath 或绝对路径均可），避免出错后原文件丢失
func (u *Upload) Update(targetPath string, compress bool) *Upload {
	if u.Error != nil {
		return u
	}
	originPath := targetPath

	// 安全判断：不能包含.. 防止路径穿越到项目外
	if strings.Contains(targetPath, "..") {
		u.Error = errors.New("路径不合法，不能包含..")
		return u
	}

	// 路径不能为空
	if targetPath == "" {
		u.Error = errors.New("目标路径不能为空")
		return u
	}
	// targetPath = strings.TrimLeft(strings.TrimPrefix(strings.TrimPrefix(targetPath, "/"), u.config.BasePath), "\\/")

	// 完整目标路径
	// dst := u.config.BasePath + "/" + targetPath
	if !u.config.PathHasBase {
		targetPath = "/" + u.config.BasePath + "/" + targetPath
	}
	targetPath = strings.TrimPrefix(targetPath, "/")

	// 确保目标目录存在
	if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		u.Error = err
		return u
	}

	// 先保存到临时路径
	tempPath := targetPath + ".tmp." + fmt.Sprintf("%d", time.Now().UnixNano())

	fileReader, err := u.fileHeader.Open()
	if err != nil {
		u.Error = err
		return u
	}
	defer fileReader.Close()

	destFile, err := os.Create(tempPath)
	if err != nil {
		u.Error = err
		return u
	}
	defer destFile.Close()

	saved := false
	if compress && u.File.FileType == "image" {
		fileReader.Seek(0, io.SeekStart)
		if img, kind, err := image.Decode(fileReader); err == nil {
			switch kind {
			case "jpeg":
				err = jpeg.Encode(destFile, img, &jpeg.Options{Quality: 80})
			case "png":
				err = png.Encode(destFile, img)
			}
			if err == nil {
				saved = true
				if info, err := destFile.Stat(); err == nil {
					u.File.Size = info.Size()
				}
			}
		}
	}

	if !saved {
		fileReader.Seek(0, io.SeekStart)
		_, err = io.Copy(destFile, fileReader)
		if err != nil {
			os.Remove(tempPath)
			u.Error = err
			return u
		}
	}
	destFile.Close()

	// 删除旧文件（放最后删除，避免出错后原文件丢失）
	if err := os.Remove(targetPath); err != nil && !os.IsNotExist(err) {
		os.Remove(tempPath)
		u.Error = err
		return u
	}

	// 确保目标目录存在（Windows 上 os.Rename 要求目录必须存在，且不能跨卷）
	if err := os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
		os.Remove(tempPath)
		u.Error = err
		return u
	}


	// 移动临时文件到目标路径
	if err := os.Rename(tempPath, targetPath); err != nil {
		os.Remove(tempPath)
		u.Error = err
		return u
	}
	u.File.Path = originPath
	if u.config.Domain != "" {
		u.File.Url = u.config.Domain + originPath
	}

	return u
}

func (u *Upload) SetDir(dir string) *Upload {
	if u.Error != nil || dir == "" {
		return u
	}

	if len(dir) > 200 { // 路径长度限制
		u.Error = errors.New("路径过长")
		return u
	}

	u.Dir = strings.TrimLeft(strings.TrimPrefix(strings.TrimLeft(dir, "\\/"), u.config.BasePath), "\\/")
	if !strings.HasPrefix(u.Dir, UPLOAD_DIR) {
		u.Dir = UPLOAD_DIR + "/" + u.Dir
	}

	if strings.Contains(u.Dir, "..") || filepath.IsAbs(u.Dir) {
		u.Error = errors.New("无效的路径")
		return u
	}
	return u
}

func (u *Upload) VerifType(t string) *Upload {
	if u.Error != nil || t == "" {
		return u
	}

	if (t == "image" || t == "video") && t != u.File.FileType {
		u.Error = errors.New("文件类型不匹配")
		return u
	}

	return u
}

func (u *Upload) Save(compress bool) *Upload {
	if u.Error != nil {
		return u
	}

	name := uuid.New().String()
	dst := u.config.BasePath + "/" + u.Dir + "/" + name
	if u.File.Extension != "" {
		dst += "." + u.File.Extension
	}
	if u.config.PathHasBase {
		u.File.Path = "/" + dst
	} else {
		u.File.Path = fmt.Sprintf("/%s/%s", u.Dir, name)
		if u.File.Extension != "" {
			u.File.Path += "." + u.File.Extension
		}
	}

	if u.config.Domain != "" {
		u.File.Url = u.config.Domain + u.File.Path
	}

	// 创建文件目录
	if err := os.MkdirAll(filepath.Dir(dst), os.ModePerm); err != nil {
		u.Error = err
		return u
	}

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

	saved := false

	// 压缩图片
	if compress && u.File.FileType == "image" {
		fileReader.Seek(0, io.SeekStart) // 需要重置文件指针到开头
		if img, kind, err := image.Decode(fileReader); err == nil {
			switch kind {
			case "jpeg":
				err = jpeg.Encode(destFile, img, &jpeg.Options{Quality: 80})
			case "png":
				err = png.Encode(destFile, img)
			}
			if err == nil {
				saved = true
				if info, err := destFile.Stat(); err == nil {
					u.File.Size = info.Size()
				}
			}
		}
	}

	if !saved {
		fileReader.Seek(0, io.SeekStart)
		_, err = io.Copy(destFile, fileReader)
		if err != nil {
			u.Error = err
			return u
		}
	}

	return u
}

func (u *Upload) Limit() *Upload {
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
		if u.File.Size > u.config.MaxImageSize*1024*1024 {
			u.Error = fmt.Errorf("图片不能超过%dM", u.config.MaxImageSize)
			return u
		}
	case "video":
		if u.File.Size > u.config.MaxVideoSize*1024*1024 {
			u.Error = fmt.Errorf("视频不能超过%dM", u.config.MaxVideoSize)
			return u
		}
	case "other":
		if u.File.Size > u.config.MaxOtherSize*1024*1024 {
			u.Error = fmt.Errorf("文件不能超过%dM", u.config.MaxOtherSize)
			return u
		}
	}

	if !slices.Contains(u.config.Extensions, u.File.Extension) {
		u.Error = fmt.Errorf("%s格式不允许上传", u.File.Extension)
		return u
	}

	return u
}

func (u *Upload) GetFile() (*File, error) {
	if u.Error != nil {
		return nil, u.Error
	}
	return u.File, nil
}

func (u *Upload) detectMimeType() error {
	fileReader, err := u.fileHeader.Open()
	if err != nil {
		return err
	}
	defer fileReader.Close()
	buffer := make([]byte, 512)
	_, err = fileReader.Read(buffer)
	if err != nil {
		return err
	}
	fileReader.Seek(0, io.SeekStart)
	u.File.Mime = http.DetectContentType(buffer)
	return nil
}
