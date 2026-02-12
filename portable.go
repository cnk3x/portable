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

	"github.com/adrg/xdg"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/tidwall/jsonc"
	"github.com/valyala/fasttemplate"
	"go.yaml.in/yaml/v4"
)

const CONFIG_FILE_NAME = "portable.yaml"
const CONFIG_NAME = "portable"

// LoadApps finds the directory that contains a portable.yaml file.
func LoadApps(root string, maxDepth int) (apps []*PortableApp) {
	root, _ = filepath.Abs(root)

	filepath.WalkDir(root, func(path string, d fs.DirEntry, e error) (err error) {
		if err = e; err != nil {
			slog.Debug("读取路径失败", "path", path, "err", err)
			return nil
		}

		if !d.IsDir() {
			return nil
		}

		if strings.HasPrefix(d.Name(), ".") || strings.HasPrefix(d.Name(), "_") || strings.HasPrefix(d.Name(), "~") {
			return fs.SkipDir
		}

		if rel, _ := filepath.Rel(root, path); rel != "." {
			depth := strings.Count(rel, string(filepath.Separator)) + 1
			if depth > maxDepth {
				return fs.SkipDir
			}
		}

		if d.IsDir() {
			if app, e := LoadApp(path); e == nil {
				apps = append(apps, app)
			}
		}
		return
	})
	return
}

// LoadApp load app config
func LoadApp(dir string) (app *PortableApp, err error) {
	app = &PortableApp{}
	app.root, _ = filepath.Abs(dir)

	for _, ext := range []string{".json", ".jsonc", ".yaml", ".yml"} {
		app.configPath = filepath.Join(app.root, CONFIG_NAME+ext)
		if err = unmarshalFile(app.configPath, app); !os.IsNotExist(err) {
			return
		}
	}

	if err != nil {
		return
	}

	if app.Name == "" {
		app.Name = filepath.Base(app.root)
	}

	slog.Info("binds", "name", app.Name, "len", len(app.Bind))

	// xDictRepl := newXDictRepl(xDict)

	// for i := range app.Bind {
	// 	slog.Info(app.Name, "name1", app.Bind[i].Name)
	// 	app.Bind[i].Name = app.Abs(xDictRepl(app.Bind[i].Name))
	// 	slog.Info(app.Name, "name2", app.Bind[i].Name)
	// 	slog.Info(app.Name, "Target1", app.Bind[i].Target)
	// 	app.Bind[i].Target = app.Abs(xDictRepl(app.Bind[i].Target))
	// 	slog.Info(app.Name, "Target2", app.Bind[i].Target)
	// }

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
	slog.Info("install", "name", app.Name)
	xDictRepl := newXDictRepl(app.root)

	for _, b := range app.Bind {
		b.Name = xDictRepl(b.Name)
		b.Target = xDictRepl(b.Target)
		b.Name = app.Abs(b.Name)
		b.Target = app.Abs(b.Target)

		slog.Info("   bind", "action", b.Action, "name", b.Name, "target", b.Target)
		if !dirtyRun {
			if e := b.Create(); e != nil {
				slog.Error("      -", "err", e)
				err = errors.Join(err, e)
			}
		}
	}
	return
}

// Uninstall the app
func (app *PortableApp) Uninstall(dirtyRun bool) (err error) {
	slog.Info("uninstall", "name", app.Name)
	for _, b := range app.Bind {
		slog.Info("   remove", "name", b.Name)
		slog.Info("        -", "action", "remove")
		slog.Info("        -", "target", b.Target)
		if e := os.Remove(b.Name); e != nil && !os.IsNotExist(e) {
			slog.Error("        -", "err", e)
			err = errors.Join(err, e)
		}
	}
	return
}

func (b Bind) Create() (err error) {
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

// Create shortcuts
func createShortcut(target, name string) (err error) {
	if name == "" || target == "" {
		return fmt.Errorf("name or target is empty")
	}

	if len(name) <= 4 || name[len(name)-4:] != ".lnk" {
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

// newXDictRepl: 变量替换
func newXDictRepl(root string) func(s string) string {
	var (
		base    = root
		roaming = cmp.Or(xdg.DataDirs...)
		start   = filepath.Join(roaming, "Microsoft", "Windows", "Start Menu", "PortableApps")
		desktop = xdg.UserDirs.Desktop
		// dataDir = filepath.Join(base, "Data")
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

// func iif[T any](c bool, t, f T) T {
// 	if c {
// 		return t
// 	}
// 	return f
// }
