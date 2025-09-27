package file

type File struct {
	Name      string `json:"name"`      // 文件名，不包含扩展名
	Extension string `json:"extension"` // 扩展名
	FullName  string `json:"full_name"` // 文件名，包含扩展名
	Path      string `json:"path"`      // 文件保存路径
	Url       string `json:"url"`       // 文件访问路径
	Size      int64  `json:"size"`      // 文件大小
	Type      string `json:"type"`      // 文件类型：image、video、other、dir
	IsDir     bool   `json:"is_dir"`    // 是否是文件夹
	ModTime   string `json:"mod_time"`  // 修改时间
	Mime      string `json:"mime"`      // 文件mime
}

type FileConfig struct {
	Domain string `json:"domain" toml:"domain" yaml:"domain"` // 域名
	Static string `json:"static" toml:"static" yaml:"static"` // 静态文件目录
}
