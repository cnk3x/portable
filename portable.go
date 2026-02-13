package portable

import (
	"cmp"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/adrg/xdg"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/tidwall/jsonc"
	"github.com/valyala/fasttemplate"
	"go.yaml.in/yaml/v4"
)

const CONFIG_FILE_NAME = "portable.yaml"
const CONFIG_NAME = "portable"

// FindApps finds the directory that contains a portable.yaml file.
func FindApps(roots []string, maxDepth int) (apps []*PortableApp) {
	loadApp := func(path string) (app PortableApp, err error) {
		app.root, _ = filepath.Abs(path)

		for _, ext := range []string{".json", ".jsonc", ".yaml", ".yml"} {
			app.configPath = filepath.Join(app.root, CONFIG_NAME+ext)
			if err = unmarshalFile(app.configPath, &app); !os.IsNotExist(err) {
				break
			}
		}

		if err != nil {
			return
		}

		if app.Name == "" {
			app.Name = filepath.Base(app.root)
		}

		for _, b := range app.Bind {
			b.Name, b.Target = app.xRepl(b.Name, b.Action == "shortcut"), app.xRepl(b.Target, false)
		}

		return
	}

	if maxDepth == 0 {
		for _, root := range roots {
			app, err := loadApp(root)
			if err != nil {
				continue
			}
			apps = append(apps, &app)
		}
		return
	}

	for _, root := range roots {
		root, _ = filepath.Abs(root)
		maxDepth = max(maxDepth, 1)

		filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				slog.Warn("读取路径失败", "path", path, "err", toErrno(err))
				return nil
			}

			if !d.IsDir() {
				return nil
			}

			if strings.HasPrefix(d.Name(), ".") || strings.HasPrefix(d.Name(), "_") || strings.HasPrefix(d.Name(), "~") {
				return fs.SkipDir
			}

			depth := 0
			if rel, _ := filepath.Rel(root, path); rel != "." {
				if depth = strings.Count(rel, string(filepath.Separator)) + 1; depth > maxDepth {
					return fs.SkipDir
				}
			}

			app, err := loadApp(path)
			if err != nil {
				slog.Debug("check dir", "path", path, "depth", depth, "max_depth", maxDepth, "name", app.Name, "err", toErrno(err))
				return nil
			}
			slog.Info("find app", "name", app.Name)

			apps = append(apps, &app)
			return nil
		})
	}
	return
}

// PortableApp struct
type PortableApp struct {
	Name string  `json:"name"`
	Bind []*Bind `json:"bind"`

	configPath string
	root       string
}

type Bind struct {
	Action string `json:"action"`
	Name   string `json:"name"`
	Target string `json:"target"`
}

func (app *PortableApp) Abs(path string) string {
	if filepath.IsAbs(path) {
		return filepath.Clean(path)
	}
	return filepath.Join(app.root, path)
}

func (app *PortableApp) ConfigPath() string { return app.configPath }

// Install the app
func (app *PortableApp) Install(dirtyRun bool) (err error) {
	slog.Info(app.Name, "action", "install", "dirty", dirtyRun)
	for _, b := range app.Bind {
		slog.Info(app.Name, "action", b.Action, "name", b.Name, "target", b.Target)
		if !dirtyRun {
			if e := b.create(); e != nil {
				slog.Error(app.Name, "action", b.Action, "err", toErrno(e))
				err = errors.Join(err, e)
			}
		}
	}
	return
}

// Uninstall the app
func (app *PortableApp) Uninstall(dirtyRun bool) (err error) {
	slog.Info(app.Name, "action", "uninstall", "dirty", dirtyRun)
	for _, b := range app.Bind {
		slog.Info(app.Name, "action", "remove", "path", b.Name)
		if !dirtyRun {
			if e := os.Remove(b.Name); e != nil {
				slog.Error(app.Name, "action", "remove", "err", toErrno(e))
				err = errors.Join(err, e)
			}
		}
	}
	return
}

func (app *PortableApp) xRepl(path string, isShortcut bool) string {
	path = fasttemplate.ExecuteFuncString(path, "${", "}", func(w io.Writer, tag string) (int, error) {
		if tag = strings.TrimSpace(tag); tag != "" {
			if tag == "base" {
				return w.Write([]byte(app.root))
			}
			if i := slices.IndexFunc(xDict, func(item [2]string) bool { return strings.EqualFold(tag, item[0]) }); i >= 0 {
				return w.Write([]byte(xDict[i][1]))
			}
		}
		return 0, nil
	})

	if filepath.IsAbs(path) {
		path = filepath.Clean(path)
	} else {
		path = filepath.Join(app.root, path)
	}

	if isShortcut {
		if len(path) <= 4 || !strings.EqualFold(path[len(path)-4:], ".lnk") {
			path += ".lnk"
		}
	}

	return path
}

func (b Bind) create() (err error) {
	if err = os.MkdirAll(filepath.Dir(b.Name), 0755); err != nil {
		return
	}
	if err = os.Remove(b.Name); err != nil && !os.IsNotExist(err) {
		return
	}
	switch b.Action {
	case "shortcut":
		return createShortcut(b.Target, b.Name)
	default:
		return os.Symlink(b.Target, b.Name)
	}
}

// Create shortcut
func createShortcut(target, name string) (err error) {
	if name == "" || target == "" {
		return fmt.Errorf("name or target is empty")
	}

	if len(name) <= 4 || !strings.EqualFold(name[len(name)-4:], ".lnk") {
		name += ".lnk"
	}

	workdir := filepath.Dir(target)

	// runtime.LockOSThread()
	// defer runtime.UnlockOSThread()
	//
	if err = ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY); err != nil {
		return
	}
	defer ole.CoUninitialize()

	var oleObj *ole.IUnknown
	if oleObj, err = oleutil.CreateObject("WScript.Shell"); err != nil {
		return
	}
	defer oleObj.Release()

	ws, e := oleObj.QueryInterface(ole.IID_IDispatch)
	if err = e; err != nil {
		return
	}
	defer ws.Release()

	oleV, e := oleutil.CallMethod(ws, "CreateShortcut", name)
	if err = e; err != nil {
		return
	}

	disp := oleV.ToIDispatch()
	defer disp.Release()

	// // Shortcut is a shortcut.
	// type Shortcut struct {
	// 	Name        string `json:"name"`
	// 	Target      string `json:"target"`
	// 	Workdir     string `json:"workdir"`
	// 	Args        string `json:"args"`
	// 	Description string `json:"description"`
	// 	Hotkey      string `json:"hotkey"`
	// 	Style       string `json:"style"`
	// 	Icon        string `json:"icon"`
	// }

	for _, props := range [][2]string{
		{"TargetPath", target},
		{"WorkingDirectory", workdir},
		// {"Arguments", shortcut.Args},
		// {"Description", shortcut.Description},
		// {"Hotkey", shortcut.Hotkey},
		// {"WindowStyle", shortcut.Style},
		// {"IconLocation", shortcut.Icon},
	} {
		if k, v := props[0], props[1]; v != "" {
			if _, err = oleutil.PutProperty(disp, k, v); err != nil {
				return
			}
		}
	}

	_, err = oleutil.CallMethod(disp, "Save")
	return
}

func unmarshalFile(path string, v any) (err error) {
	data, e := os.ReadFile(path)
	if err = e; err != nil {
		return
	}

	switch ext := strings.ToLower(filepath.Ext(path)); ext {
	case ".json", ".jsonc":
		data = jsonc.ToJSONInPlace(data)
	case ".yaml", ".yml":
		var tmp any
		if err = yaml.Unmarshal(data, &tmp); err != nil {
			return
		}
		if data, err = json.Marshal(tmp); err != nil {
			return
		}
	default:
		err = fmt.Errorf("unsupported file type: %s", ext)
		return
	}

	if err = json.Unmarshal(data, v); err != nil {
		return
	}

	if rawSet, ok := v.(interface{ setRaw(raw json.RawMessage) }); ok {
		rawSet.setRaw(data)
	}
	return
}

var xDict = [][2]string{
	{"Local", xdg.DataHome},                       // D:\Users\user\AppData\Local
	{"Roaming", cmp.Or(xdg.DataDirs...)},          // D:\Users\user\AppData\Roaming
	{"Home", xdg.Home},                            // D:\Users\user
	{"Programs", xdg.BinHome},                     // D:\Users\user\AppData\Local\Programs
	{"ProgramFiles", `C:\Program Files`},          // C:\Program Files
	{"ProgramFilesX86", `C:\Program Files (x86)`}, // C:\Program Files (x86)
	{"Desktop", xdg.UserDirs.Desktop},             // D:\Desktop
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

	// D:\Users\user\AppData\Roaming\Microsoft\Windows\Start Menu\PortableApps
	{"Start", filepath.Join(cmp.Or(xdg.DataDirs...), "Microsoft", "Windows", "Start Menu", "PortableApps")},
}

func toErrno(err error) error {
	if err == nil {
		return nil
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		err = errno
	}
	return err
}
