package portable

import (
	"cmp"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/adrg/xdg"
	"github.com/cnk3x/portable/pkg/files"
	"github.com/cnk3x/portable/pkg/it"
	"github.com/cnk3x/portable/pkg/shortcuts"
	"github.com/goccy/go-yaml"
	"github.com/valyala/fasttemplate"
)

const CONFIG_FILE_NAME = "portable.yaml"

// FindDirs finds the directory that contains a portable.yaml file.
func FindDirs(root string, depth int) (dirs []string) {
	_ = files.WalkDir(root, func(path string, d fs.DirEntry, e error) (err error) {
		if err = e; err != nil {
			return
		}

		if d.IsDir() {
			if stat, e := os.Stat(filepath.Join(path, CONFIG_FILE_NAME)); e == nil && stat.Mode().IsRegular() {
				dirs = append(dirs, path)
			}
		}

		return
	}, depth)
	return
}

// PortableApp struct
type PortableApp struct {
	Name     string                `json:"name"`
	Bind     []string              `json:"bind"`
	Shortcut []*shortcuts.Shortcut `json:"shortcut"`

	configPath string
	binds      [][2]string //[source, link]
}

// LoadApp load app config
func LoadApp(configPath string) (app *PortableApp, err error) {
	app = &PortableApp{configPath: configPath}

	if app.configPath, err = filepath.Abs(cmp.Or(app.configPath, ".")); err != nil {
		return
	}

	if s, e := os.Stat(app.configPath); e == nil && s.IsDir() {
		app.configPath = filepath.Join(app.configPath, CONFIG_FILE_NAME)
	}

	var data []byte
	if data, err = os.ReadFile(app.configPath); err != nil {
		return
	}

	if err = yaml.UnmarshalWithOptions(data, app, yaml.UseJSONUnmarshaler()); err != nil {
		return
	}

	if app.Name == "" {
		app.Name = filepath.Base(filepath.Dir(app.configPath))
	}

	var (
		base    = filepath.Dir(app.configPath)
		dataDir = filepath.Join(base, "Data")
		roaming = cmp.Or(xdg.DataDirs...)
		start   = filepath.Join(roaming, "Microsoft", "Windows", "Start Menu", "PortableApps")
		desktop = xdg.UserDirs.Desktop
	)

	xDict := [][2]string{
		{"Local", xdg.DataHome},                       // D:\Users\user\AppData\Local
		{"Roaming", roaming},                          // D:\Users\user\AppData\Roaming
		{"Home", xdg.Home},                            // D:\Users\user
		{"Programs", xdg.BinHome},                     // D:\Users\user\AppData\Local\Programs
		{"ProgramFiles", `C:\Program Files`},          // C:\Program Files
		{"ProgramFilesX86", `C:\Program Files (x86)`}, // C:\Program Files (x86)
		{"Desktop", desktop},                          // D:\Desktop
		{"Document", xdg.UserDirs.Documents},          // D:\Documents
		{"Documents", xdg.UserDirs.Documents},         // D:\Documents
		{"Download", xdg.UserDirs.Download},           // D:\Downloads
		{"Downloads", xdg.UserDirs.Download},          // D:\Downloads
		{"Music", xdg.UserDirs.Music},                 // D:\Music
		{"Musics", xdg.UserDirs.Music},                // D:\Music
		{"Picture", xdg.UserDirs.Pictures},            // D:\Pictures
		{"Pictures", xdg.UserDirs.Pictures},           // D:\Pictures
		{"Video", xdg.UserDirs.Videos},                // D:\Videos
		{"Videos", xdg.UserDirs.Videos},               // D:\Videos
		{"Start", start},                              // D:\Users\user\AppData\Roaming\Microsoft\Windows\Start Menu\PortableApps
		{"Base", base},
	}

	// 倒序
	slices.SortStableFunc(xDict, func(a, b [2]string) int { return strings.Compare(b[1], a[1]) })

	xDictRepl := newXDictRepl(xDict)

	// 快捷方式
	for i, item := range app.Shortcut {
		app.Shortcut[i].Name = xDictRepl(item.Name)
		app.Shortcut[i].Target = xDictRepl(item.Target)
		app.Shortcut[i].Workdir = xDictRepl(item.Workdir)
		app.Shortcut[i].Args = xDictRepl(item.Args)

		if len(item.Place) > 0 {
			it.ForIndex(item.Place, func(j int) { app.Shortcut[i].Place[j] = xDictRepl(item.Place[j]) })
		} else {
			app.Shortcut[i].Place = []string{base, desktop, filepath.Join(start, item.Category)}
		}
	}

	// 绑定路径
	targetPath := shortPath(xDict, dataDir)

	// 绑定目录
	for _, target := range app.Bind {
		target = xDictRepl(target)
		app.binds = append(app.binds, [2]string{targetPath(target), target})
	}

	return
}

// ConfigPath config file path
func (app *PortableApp) ConfigPath() string { return app.configPath }

// Install the app
func (app *PortableApp) Install(force bool) (err error) {
	err = errors.Join(err, createSymlinks(app.binds, force))     // 绑定目录
	err = errors.Join(err, createShortcuts(app.Shortcut, force)) // 创建快捷方式
	return
}

// Uninstall the app
func (app *PortableApp) Uninstall(force bool) (err error) {
	err = errors.Join(err, removeSymlinks(app.binds, force))     // 移除绑定链接
	err = errors.Join(err, removeShortcuts(app.Shortcut, force)) // 移除快捷方式
	return
}

// createSymlinks create symlinks
func createSymlinks(symlinks [][2]string, force bool) (err error) {
	for _, symlink := range symlinks {
		source, link := symlink[0], symlink[1]
		e := files.CreateSymlink(source, link, force, true)
		if e != nil {
			slog.Error("linkbind", "link", link)
			slog.Error("       -", "source", source)
			slog.Error("       -", "err", e)
		} else {
			slog.Debug("linkbind", "link", link)
			slog.Debug("       -", "source", source)
		}
		err = errors.Join(err, e)
	}
	return
}

// removeSymlinks remove symlinks
func removeSymlinks(symlinks [][2]string, force bool) (err error) {
	for _, symlink := range symlinks {
		link := symlink[1]
		e := files.RemoveSymlink(link, force)
		if err = errors.Join(err, e); e != nil {
			slog.Error("remove linkbind", "path", link)
			slog.Error("               -", "err", e)
		} else {
			slog.Debug("remove linkbind", "path", link)
		}
	}
	return
}

// createShortcuts create shortcuts
func createShortcuts(items []*shortcuts.Shortcut, _ bool) (err error) {
	return shortcuts.Create(items, func(shortcut *shortcuts.Shortcut, path string, err error) {
		if err != nil {
			slog.Error("shortcut", "Path", path)
			slog.Error("       -", "Target", shortcut.Target)
			if shortcut.Args != "" {
				slog.Error("       -", "Args", shortcut.Args)
			}
			slog.Error("       -", "Workdir", shortcut.Workdir)
			slog.Error("       -", "err", err)
		} else {
			slog.Debug("shortcut", "Path", path)
			slog.Debug("       -", "Target", shortcut.Target)
			if shortcut.Args != "" {
				slog.Debug("       -", "Args", shortcut.Args)
			}
			slog.Debug("       -", "Workdir", shortcut.Workdir)
		}
	})
}

// removeShortcuts remove shortcuts
func removeShortcuts(items []*shortcuts.Shortcut, force bool) (err error) {
	return shortcuts.Remove(items, func(shortcut *shortcuts.Shortcut, path string, e error) {
		if e != nil {
			slog.Error("remove shortcut", "path", path)
			slog.Error("               -", "err", e)
		} else {
			slog.Debug("remove shortcut", "path", path)
		}
	}, force)
}

// shortPath 替换路径为前缀代号
func shortPath(xDict [][2]string, root string) func(fullPath string) string {
	return func(fullPath string) string {
		var name, rel string
		if i := slices.IndexFunc(xDict, func(item [2]string) bool { return hasPrefixFold(fullPath, item[1]) }); i >= 0 {
			name, rel = xDict[i][0], strings.Trim(fullPath[len(xDict[i][1]):], `\/`)
		} else {
			dir, fn := filepath.Split(fullPath)
			name, rel = filepath.Base(dir), fn
		}
		return filepath.Join(root, name, rel)
	}
}

// newXDictRepl: 变量替换
func newXDictRepl(xDict [][2]string) func(s string) string {
	tagFind := func(tag string) func(item [2]string) bool {
		return func(item [2]string) bool { return strings.EqualFold(tag, item[0]) }
	}

	return func(s string) string {
		return fasttemplate.ExecuteFuncString(s, "${", "}", func(w io.Writer, tag string) (int, error) {
			if tag = strings.TrimSpace(tag); tag != "" {
				if i := slices.IndexFunc(xDict, tagFind(tag)); i >= 0 {
					return w.Write([]byte(xDict[i][1]))
				}
			}
			return 0, nil
		})
	}
}

// hasPrefixFold: 不区分大小写的 strings.HasPrefix
func hasPrefixFold(s string, prefix string) bool {
	return len(s) >= len(prefix) && strings.EqualFold(s[0:len(prefix)], prefix)
}
