package portable

import (
	"cmp"
	"errors"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/adrg/xdg"
	"github.com/cnk3x/portable/pkg/files"
	"github.com/cnk3x/portable/pkg/shortcuts"
	"github.com/valyala/fasttemplate"
	"gopkg.in/yaml.v3"
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

	if err = yaml.Unmarshal(data, app); err != nil {
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

	xdgTags := [][2]string{
		{"local", xdg.DataHome},               // D:\Users\user\AppData\Local
		{"roaming", roaming},                  // D:\Users\user\AppData\Roaming
		{"home", xdg.Home},                    // D:\Users\user
		{"programs", xdg.BinHome},             // D:\Users\user\AppData\Local\Programs
		{"desktop", desktop},                  // D:\Desktop
		{"document", xdg.UserDirs.Documents},  // D:\Documents
		{"documents", xdg.UserDirs.Documents}, // D:\Documents
		{"download", xdg.UserDirs.Download},   // D:\Downloads
		{"downloads", xdg.UserDirs.Download},  // D:\Downloads
		{"music", xdg.UserDirs.Music},         // D:\Music
		{"musics", xdg.UserDirs.Music},        // D:\Music
		{"picture", xdg.UserDirs.Pictures},    // D:\Pictures
		{"pictures", xdg.UserDirs.Pictures},   // D:\Pictures
		{"video", xdg.UserDirs.Videos},        // D:\Videos
		{"videos", xdg.UserDirs.Videos},       // D:\Videos
		{"start", start},                      // D:\Users\user\AppData\Roaming\Microsoft\Windows\Start Menu\PortableApps
		{"base", base},
	}

	findTag := func(path string) string {
		for _, v := range xdgTags {
			if strings.HasPrefix(path, v[1]) {
				rel := strings.TrimPrefix(path, v[1])
				name := strings.ToUpper(v[0][:1]) + v[0][1:]
				return filepath.Join(name, rel)
			}
		}

		dir, name := filepath.Split(path)
		return filepath.Join(filepath.Base(dir), name)
	}

	tplRepl := func() func(s string) string {
		mapTags := make(map[string]string, len(xdgTags))
		for _, v := range xdgTags {
			mapTags[v[0]] = v[1]
		}
		tagFunc := func(w io.Writer, tag string) (int, error) {
			if v := mapTags[strings.TrimSpace(strings.ToLower(tag))]; v != "" {
				return w.Write([]byte(v))
			}
			return 0, nil
		}
		return func(s string) string {
			return fasttemplate.ExecuteFuncString(s, "${", "}", tagFunc)
		}
	}()

	// 快捷方式
	for i, it := range app.Shortcut {
		app.Shortcut[i].Name = tplRepl(it.Name)
		app.Shortcut[i].Target = tplRepl(it.Target)
		app.Shortcut[i].Workdir = tplRepl(it.Workdir)
		app.Shortcut[i].Args = tplRepl(it.Args)

		if len(it.Place) > 0 {
			for j, place := range it.Place {
				app.Shortcut[i].Place[j] = tplRepl(place)
			}
		} else {
			app.Shortcut[i].Place = []string{base, desktop, filepath.Join(start, it.Category)}
		}
	}

	// 绑定目录
	for i, v := range app.Bind {
		app.Bind[i] = tplRepl(v)
	}

	for _, target := range app.Bind {
		source := filepath.Join(dataDir, findTag(target))
		app.binds = append(app.binds, [2]string{source, target})
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

// create symlinks
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

// remove symlinks
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

// create shortcuts
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

// remove shortcuts
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
