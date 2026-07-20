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

// Update 更新文件：根据指定路径，先保存新文件到临时位置，
// 再删除原文件，最后将新文件移动到目标路径（防止出错后文件丢失）
func (f *Files) Update(file *multipart.FileHeader, path string, l Limit) (*File, error) {
	// 安全校验：路径不能包含".."，防止目录遍历攻击
	if strings.Contains(path, "..") {
		return nil, errors.New("路径包含非法字符")
	}

	// 解析文件信息
	extension := filepath.Ext(file.Filename)
	baseName := file.Filename
	if len(extension) > 0 {
		extension = extension[1:]
		baseName = file.Filename[:len(file.Filename)-len(extension)-1]
	}
	mime := file.Header.Get("Content-Type")
	types, _, _ := strings.Cut(mime, "/")
	if types != "image" && types != "video" {
		types = "other"
	}

	// 上传限制
	err := limit(extension, types, file.Size, l, baseName)
	if err != nil {
		return nil, err
	}

	// 构造完整文件系统路径
	fullPath := f.Config.Static + "/" + strings.TrimPrefix(path, "/")

	// 确保目标目录存在
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return nil, err
	}

	// 先保存新文件到临时路径（防止原文件删除后新文件保存失败）
	tempPath := fullPath + ".tmp"
	size, err := Save(file, tempPath, l.Compress)
	if err != nil {
		return nil, err
	}

	// 删除原文件（如果存在）
	if _, statErr := os.Stat(fullPath); statErr == nil {
		if err := os.Remove(fullPath); err != nil {
			os.Remove(tempPath)
			return nil, err
		}
	}

	// 将临时文件重命名为目标路径
	if err := os.Rename(tempPath, fullPath); err != nil {
		return nil, err
	}

	// 获取文件访问域名
	base, err := Domain(f.Request, f.Config.Domain)
	if err != nil {
		return nil, err
	}

	relativePath := strings.TrimPrefix(fullPath, f.Config.Static)
	return &File{
		Name:      baseName,
		Extension: extension,
		FullName:  file.Filename,
		Path:      "/" + relativePath,
		Url:       base + "/" + relativePath,
		Size:      size,
		Type:      types,
		IsDir:     false,
		ModTime:   time.Now().Format("2006-01-02 15:04:05"),
		Mime:      mime,
	}, nil
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
