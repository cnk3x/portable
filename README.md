# mkp

auto make symlinks and shortcuts with 'portable.yaml' file for your application on windows

a portable.yaml file looks like this:

```yaml
bind:
  - '${Roaming}\my_app'
  - '${Document}\my_app'

shortcut:
  - name: My App
    target: '${Roaming}\App\my_app.exe'
    workdir: '${Roaming}\App'
    args: "--hello world"
    category: "My Apps"
```

you can use the following variables:

- `${roaming}` - roaming directory `like: C:\Users\user\AppData\Roaming`
- `${local}` - local directory `like:  C:\Users\user\AppData\Local`
- `${home}` - home directory `like:  C:\Users\user`
- `${programs}` - programs directory `like: D:\Users\user\AppData\Local\Programs`
- `${document}` - document directory `like: C:\Users\user\Documents`
- `${desktop}` - desktop directory `like: C:\Users\user\Desktop`
- `${start}` - the start menu directory `like: D:\Users\user\AppData\Roaming\Microsoft\Windows\Start Menu\PortableApps`
- `${base}` - the directory where the portable.yaml file is located
- more variables can find in [portable.go](portable.go)

install

```bash
> go build -trimpath -ldflags '-s -w' -v github.com/cnk3x/portable/cmd/mkp

> mkp -h
manage portable app

Usage:
  mkp.exe [command]

Available Commands:
  install     install portable app
  list        list portable app
  uninstall   uninstall portable app

Flags:
  -d, --debug     debug output
  -h, --help      help for mkp.exe
  -v, --verbose   verbose output

Use "mkp.exe [command] --help" for more information about a command.
```
