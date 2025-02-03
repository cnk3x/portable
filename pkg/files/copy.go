package files

type CopyOptions struct {
	RemoveSource bool                              // 是否删除源文件（移动）
	OnExist      func(source, target string) error // 当目标存在时，如何处理
	OnSymlink    func(source, target string) error // 当源为软链接时，如何处理
	Error        func(err error) error
	Symlink      bool
}

func CopyFile(source, target string) (err error) {
	var crossDevice bool
	_ = crossDevice
	// 跨驱动器模式, 只能流式读取写入删除
	//

	return
}

func OverwriteOnExist(source, target string) error { return nil }

func SkipOnExist(source, target string) error { return nil }
