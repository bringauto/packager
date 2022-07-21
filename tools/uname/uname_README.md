
# Fake uname utility

Fake `uname` utility reads data from config file nd
provide them to the user.

Config file has a name `uname.txt` and must be locatad at the same
directory as the fake `uname`

To generate config file just run **real** `uname` utility as

```
uname -a > uname.txt
```

Update the `uname.txt` by your needs.

## Supported flags

- **-m** - machine

## Unsupported flags

- **-c**
- **-a**
- **-v**
- **-h**
- **-d**

## Build

Run command:

- `CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -tags netgo -ldflags '-w'  `

