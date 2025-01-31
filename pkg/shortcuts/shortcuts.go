package shortcuts

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/cnk3x/portable/pkg/files"
	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

type (
	CompletedFunc = func(shortcut *Shortcut, path string, err error)

	oleShellFunc = func(ws *ole.IDispatch) error
)

// Shortcut is a shortcut.
type Shortcut struct {
	Name        string `json:"name"`
	Target      string `json:"target"`
	Workdir     string `json:"workdir"`
	Args        string `json:"args"`
	Description string `json:"description"`
	Hotkey      string `json:"hotkey"`
	Style       string `json:"style"`
	Icon        string `json:"icon"`

	Category string   `json:"category"` // 当Place为空时，Category 为开始菜单的子菜单
	Place    []string `json:"place"`
}

// Create shortcuts
func Create(shortcuts []*Shortcut, onCompleted CompletedFunc) error {
	var calls []oleShellFunc
	for _, shortcut := range shortcuts {
		calls = append(calls, makeCall(shortcut, onCompleted))
	}
	return useWsShell(calls...)
}

// Remove shortcuts
func Remove(shortcuts []*Shortcut, onCompleted CompletedFunc, force bool) (err error) {
	// 移除快捷方式
	for _, shortcut := range shortcuts {
		for _, place := range shortcut.Place {
			path := filepath.Join(place, shortcut.Name+".lnk")
			e := files.RemoveSymlink(path, force)
			if onCompleted != nil {
				onCompleted(shortcut, path, e)
			}
			err = errors.Join(err, e)
		}
	}
	return
}

func makeCall(shortcut *Shortcut, onCompleted CompletedFunc) oleShellFunc {
	return func(ws *ole.IDispatch) (err error) {
		if len(shortcut.Place) == 0 {
			return
		}

		files.AbsUpdate(&shortcut.Target, &shortcut.Workdir)

		for i, v := range shortcut.Place {
			files.AbsUpdate(&v)
			shortcut.Place[i] = v
		}

		if shortcut.Workdir == "" {
			shortcut.Workdir = filepath.Dir(shortcut.Target)
		}

		for _, place := range shortcut.Place {
			files.AbsUpdate(&place)
			path := filepath.Join(place, shortcut.Name+".lnk")

			e := os.MkdirAll(place, 0o755)
			if e == nil {
				e = createInWs(ws, shortcut, path)
			}

			if onCompleted != nil {
				onCompleted(shortcut, path, e)
			}

			err = errors.Join(err, e)
		}
		return
	}
}

func createInWs(ws *ole.IDispatch, shortcut *Shortcut, path string) (err error) {
	var oleV *ole.VARIANT
	if oleV, err = oleutil.CallMethod(ws, "CreateShortcut", path); err != nil {
		return
	}

	putProperties := map[string]any{
		"TargetPath":       shortcut.Target,
		"Arguments":        shortcut.Args,
		"WorkingDirectory": shortcut.Workdir,
		"Description":      shortcut.Description,
		"Hotkey":           shortcut.Hotkey,
		"WindowStyle":      shortcut.Style,
		"IconLocation":     shortcut.Icon,
	}

	shortcutObj := oleV.ToIDispatch()
	defer shortcutObj.Release()

	for k, v := range putProperties {
		if v != "" {
			if _, err = oleutil.PutProperty(shortcutObj, k, v); err != nil {
				return
			}
		}
	}

	if _, err = oleutil.CallMethod(shortcutObj, "Save"); err != nil {
		return
	}

	return
}

func useWsShell(calls ...oleShellFunc) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err = ole.CoInitializeEx(0, ole.COINIT_APARTMENTTHREADED|ole.COINIT_SPEED_OVER_MEMORY); err != nil {
		return
	}
	defer ole.CoUninitialize()

	var oleObj *ole.IUnknown
	if oleObj, err = oleutil.CreateObject("WScript.Shell"); err != nil {
		return
	}
	defer oleObj.Release()

	var wsShell *ole.IDispatch
	if wsShell, err = oleObj.QueryInterface(ole.IID_IDispatch); err != nil {
		return
	}
	defer wsShell.Release()

	for _, call := range calls {
		e := func(call func(*ole.IDispatch) error) (err error) {
			defer func() {
				if e := recover(); e != nil {
					if ex, ok := e.(error); ok {
						err = ex
					} else {
						err = fmt.Errorf("%v", e)
					}
				}
			}()
			err = call(wsShell)
			return
		}(call)

		if e != nil {
			err = errors.Join(err, e)
		}
	}

	return
}
