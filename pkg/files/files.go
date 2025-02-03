package files

import (
	"cmp"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// AbsUpdate 尝试将路径转换为绝对路径, 如果成功则使用绝对路径赋值
func AbsUpdate(paths ...*string) {
	for _, path := range paths {
		if *path != "" {
			if p, e := filepath.Abs(*path); e == nil {
				*path = p
			}
		}
	}
}

// WalkDir 遍历目录，同 filepath.WalkDir，支持最大层级限定(maxDepth)。
func WalkDir(root string, fn fs.WalkDirFunc, maxDepth ...int) (err error) {
	var info fs.FileInfo
	if info, err = os.Lstat(root); err != nil {
		err = fn(root, nil, err)
	} else {
		maxDepth := cmp.Or(cmp.Or(maxDepth...), -1)
		walkDir := (func(string, fs.DirEntry, int) error)(nil)
		walkDir = func(path string, d fs.DirEntry, depth int) (err error) {
			if maxDepth != -1 && depth > maxDepth {
				return nil
			}

			if err = fn(path, d, nil); err != nil || !d.IsDir() {
				if err == fs.SkipDir && d.IsDir() {
					err = nil // Successfully skipped directory.
				}
				return // 如果有错误或者非文件夹，返回
			}

			var dirs []fs.DirEntry

			if dirs, err = os.ReadDir(path); err != nil {
				// Second call, to report ReadDir error.
				if err = fn(path, d, err); err != nil {
					if err == fs.SkipDir && d.IsDir() {
						err = nil
					}
					return
				}
			}

			for _, d := range dirs {
				if err = walkDir(filepath.Join(path, d.Name()), d, depth+1); err != nil {
					if err == fs.SkipDir {
						err = nil
						break
					}
					return
				}
			}

			return
		}
		err = walkDir(root, fs.FileInfoToDirEntry(info), 0)
	}
	if err == fs.SkipDir || err == fs.SkipAll {
		err = nil
	}
	return
}

// IsSymlinkStat checks if the given file info is a symlink.
func IsSymlinkStat(stat fs.FileInfo) bool {
	return stat != nil && stat.Mode()&fs.ModeSymlink != 0
}

// IsShortcutStat checks if the file is a shortcut.
func IsShortcutStat(stat fs.FileInfo) bool {
	return stat != nil && stat.Mode().IsRegular() && strings.HasSuffix(stat.Name(), ".lnk")
}

// CreateDir 判断是否目录，如果不存在或者是软连接则删除软连接并创建目录，返回创建结果，如果不是目录，则返回目录已存在错误。
func CreateDir(dir string) (err error) {
	if stat, e := os.Lstat(dir); e != nil || stat.IsDir() {
		if os.IsNotExist(e) {
			err = os.MkdirAll(dir, 0o755)
		} else {
			err = e
		}
	} else {
		if IsSymlinkStat(stat) { // 是软链接: 删除，重建
			if err = os.Remove(dir); err == nil {
				err = os.MkdirAll(dir, 0o755)
			}
		} else {
			err = fmt.Errorf("%w and not a directory: %s", fs.ErrExist, dir)
		}
	}

	return
}

// CreateSymlink creates a symlink at path pointing to target.
func CreateSymlink(source, link string, force bool, abs bool) (err error) {
	if abs {
		AbsUpdate(&source, &link)
	}

	if err = CreateDir(filepath.Dir(link)); err != nil {
		return
	}

	if err = CreateDir(source); err != nil {
		return
	}

	// 如果是单个文件或者空文件夹，删除
	if err = RemoveSymlink(link, force); err != nil {
		return
	}

	err = os.Symlink(source, link)
	return
}

// RemoveSymlink deletes a symlink or a .lnk file, if force is true, it will delete the target file as well.
func RemoveSymlink(path string, force bool) (err error) {
	stat, e := os.Lstat(path)
	if err = e; err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
		return
	}

	if IsSymlinkStat(stat) || IsShortcutStat(stat) {
		err = os.Remove(path)
	} else if force {
		err = os.RemoveAll(path)
	} else {
		err = fmt.Errorf("%w, not a symlink and not a .lnk file (force=false)", os.ErrExist)
	}

	return
}
