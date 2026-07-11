package file

import (
	"encoding/base64"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"
)

type Files struct {
	Request *http.Request
	Config  *FileConfig
}

func NewFile(request *http.Request, cfg *FileConfig) *Files {
	return &Files{
		Request: request,
		Config:  cfg,
	}
}

// safePath 安全处理路径，防止目录穿越
func (f *Files) safePath(dir string) string {
	dir = strings.ReplaceAll(strings.TrimPrefix(dir, "/"), "/..", "")
	if !strings.HasPrefix(dir, f.Config.Static) {
		return ""
	}
	return dir
}

func (f *Files) Upload(file *multipart.FileHeader, dir string, l Limit) (*File, error) {
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

	if err := checkLimit(extension, types, file.Size, l, baseName); err != nil {
		return nil, err
	}

	path, err := f.GetPath(dir, types)
	if err != nil {
		return nil, err
	}
	path += "/" + fmt.Sprintf("%d", time.Now().UnixNano()) + "." + extension

	size, err := Save(file, path, l.Compress)
	if err != nil {
		return nil, err
	}

	base, err := Domain(f.Request, f.Config.Domain)
	if err != nil {
		return nil, err
	}

	if f.Config.NotIncludeStatic {
		path = strings.TrimPrefix(path, f.Config.Static+"/")
	}
	return &File{
		Name:      baseName,
		Extension: extension,
		FullName:  file.Filename,
		Path:      "/" + path,
		Url:       base + "/" + path,
		Size:      size,
		FileType:  types,
		Type:      types,
		IsDir:     false,
		ModTime:   time.Now().Format("2006-01-02 15:04:05"),
		Mime:      mime,
	}, nil
}

func (f *Files) Base64ToFile(b64 string, dir string, l Limit) (*File, error) {
	b64 = strings.TrimPrefix(b64, "data:")
	parts := strings.SplitN(b64, ";base64,", 2)
	types, _, _ := strings.Cut(parts[0], "/")
	extension := strings.Split(parts[0], "/")[1]
	if types != "image" && types != "video" {
		types = "other"
	}

	if err := checkLimit(extension, types, int64(len(parts[1])), l, ""); err != nil {
		return nil, err
	}

	path, err := f.GetPath(dir, types)
	if err != nil {
		return nil, err
	}
	baseName := fmt.Sprintf("%d", time.Now().UnixNano())
	path += "/" + baseName + "." + extension

	bytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	if err := os.WriteFile(path, bytes, 0644); err != nil {
		return nil, err
	}

	base, err := Domain(f.Request, f.Config.Domain)
	if err != nil {
		return nil, err
	}

	if f.Config.NotIncludeStatic {
		path = strings.TrimPrefix(path, f.Config.Static+"/")
	}
	return &File{
		Name:      baseName,
		Extension: extension,
		FullName:  baseName + "." + extension,
		Path:      "/" + path,
		Url:       base + "/" + path,
		Size:      int64(len(parts[1])),
		FileType:  types,
		Type:      types,
		IsDir:     false,
		ModTime:   time.Now().Format("2006-01-02 15:04:05"),
		Mime:      parts[0],
	}, nil
}

func (f *Files) GetPath(dir string, types string) (string, error) {
	if dir == "" {
		dir = f.Config.Static + "/upload"
		if types != "" {
			dir += "/" + types + "/" + time.Now().Format("2006-01-02")
		}
	} else {
		safe := f.safePath(dir)
		if safe == "" {
			return "", errors.New("您上传的路径不符合规范")
		}
		dir = safe
	}
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return "", err
	}
	return dir, nil
}

type Limit struct {
	ImageMaxSize int64
	VideoMaxSize int64
	OtherMaxSize int64
	Extension    string
	Compress     bool
}

func checkLimit(extension string, types string, size int64, l Limit, upName string) error {
	if size == 0 {
		return fmt.Errorf("不能上传空文件")
	}
	if len(upName) > 255 {
		return fmt.Errorf("文件名不能超过255个字符")
	}

	switch types {
	case "image":
		if l.ImageMaxSize*1024*1024 < size {
			return fmt.Errorf("图片不能超过%dM", l.ImageMaxSize)
		}
	case "video":
		if l.VideoMaxSize*1024*1024 < size {
			return fmt.Errorf("视频不能超过%dM", l.VideoMaxSize)
		}
	default:
		if l.OtherMaxSize*1024*1024 < size {
			return fmt.Errorf("文件不能超过%dM", l.OtherMaxSize)
		}
	}

	if !slices.Contains(strings.Split(l.Extension, ","), extension) {
		return fmt.Errorf("%s格式不支持上传", extension)
	}
	return nil
}

type ListParam struct {
	Path     string
	Name     string
	Page     int
	PageSize int
}
type ListRes struct {
	Data  []File `json:"data"`
	Total int64  `json:"total"`
}

func (f *Files) List(param ListParam) (*ListRes, error) {
	var err error
	param.Path, err = f.GetPath(param.Path, "")
	if err != nil {
		return nil, err
	}
	files, err := os.ReadDir(param.Path)
	if err != nil {
		return nil, err
	}

	if param.Name != "" {
		var matched []os.DirEntry
		for _, v := range files {
			if strings.Contains(v.Name(), param.Name) {
				matched = append(matched, v)
			}
		}
		files = matched
	}

	base, err := Domain(f.Request, f.Config.Domain)
	if err != nil {
		return nil, err
	}

	var list []File
	total, _, res := List(int32(param.Page), int32(param.PageSize), files)
	for _, v := range res {
		info, err := v.Info()
		if err != nil {
			return nil, err
		}

		name := v.Name()
		extension := filepath.Ext(name)
		baseName := name
		if len(extension) > 0 {
			extension = extension[1:]
			baseName = name[:len(name)-len(extension)-1]
		}

		mime, _ := FileMimeType(param.Path + "/" + name)
		types := strings.Split(mime, "/")[0]
		if types != "image" && types != "video" {
			types = "other"
		}

		path := "/" + param.Path + "/" + name
		if f.Config.NotIncludeStatic {
			path = strings.TrimPrefix(path, f.Config.Static+"/")
		}

		list = append(list, File{
			Name:      baseName,
			Extension: extension,
			FullName:  name,
			Path:      path,
			Url:       base + path,
			Size:      info.Size() / 1024,
			FileType:  types,
			Type:      types,
			IsDir:     v.IsDir(),
			ModTime:   info.ModTime().Format("2006-01-02 15:04:05"),
			Mime:      mime,
		})
	}
	return &ListRes{Data: list, Total: total}, nil
}

func (f *Files) Delete(path string, name string) error {
	if path == "" {
		path = f.Config.Static + "/upload"
	} else {
		safe := f.safePath(path)
		if safe == "" {
			return errors.New("您要删除的路径不符合规范")
		}
		path = safe
	}
	if name != "" {
		path += "/" + name
	}
	return os.RemoveAll(path)
}

func FileMimeType(path string) (string, error) {
	buffer, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return http.DetectContentType(buffer), nil
}

func Domain(r *http.Request, domain string) (string, error) {
	if domain == "" {
		ip, port, err := net.SplitHostPort(r.Host)
		if err != nil {
			return "http://" + r.Host, nil
		}
		return "http://" + ip + ":" + port, nil
	}
	return domain, nil
}

func Save(file *multipart.FileHeader, path string, compress bool) (size int64, err error) {
	size = file.Size
	fileReader, err := file.Open()
	if err != nil {
		return
	}
	defer fileReader.Close()

	destFile, err := os.Create(path)
	if err != nil {
		return
	}
	defer destFile.Close()

	if !compress {
		size, err = io.Copy(destFile, fileReader)
		return
	}

	img, kind, decodeErr := image.Decode(fileReader)
	if decodeErr != nil {
		fileReader.Seek(0, io.SeekStart)
		size, err = io.Copy(destFile, fileReader)
		return
	}

	switch kind {
	case "jpeg":
		err = jpeg.Encode(destFile, img, &jpeg.Options{Quality: 80})
	case "png":
		err = png.Encode(destFile, img)
	default:
		fileReader.Seek(0, io.SeekStart)
		size, err = io.Copy(destFile, fileReader)
		return
	}

	if err == nil {
		size, _ = destFile.Seek(0, io.SeekEnd)
	}
	return
}
