package wspad

import (
	"os"
	"path/filepath"
)

var (
	DataDir      string
	StaticDir    string
	TemplatesDir string
)

func init() {
	exe, _ := os.Executable()
	exeDir, _ := filepath.Abs(filepath.Dir(exe))
	DataDir = filepath.Join(exeDir, "db")
	StaticDir = filepath.Join(exeDir, "web/static")
	TemplatesDir = filepath.Join(exeDir, "web/templates")
}