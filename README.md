# ninixw

## これは何？

[ninix-kagari](https://github.com/Tatakinov/ninix-kagari)の精神的後継。

元プロジェクトはRuby + GTKだったのをGolang + SDLで新たに書いたもの。

## 名前の由来

Nin'i for X/Wayland/Windows

## Requirements

Required(Linux):

- go >= 1.27 (core)
- C++ compiler (gui, proxy, shm, sorakado, shiori)
- cmake (gui, proxy, shm, sorakado)
- libffi (proxy)
- fltk1.3 or fltk1.4 (gui)
- SDL3 (sorakado)
- SDL3-image (sorakado)
- SDL3-ttf (sorakado)
- jsoncpp (gui, sorakado)
- fontconfig (sorakado)
- wayland-client (sorakado)

Required(Windows):

- go >= 1.27 (core)
- C++ compiler (gui, proxy, shm, sorakado)
- cmake (gui, proxy, shm, sorakado)
- fltk1.3 (gui)
- SDL3 (sorakado)
- SDL3-image (sorakado)
- SDL3-ttf (sorakado)
- jsoncpp (gui, sorakado)

Optional(Linux/Windows):

- onnxruntime (sorakado)

## How to build

### Linux

#### Install Requirements

- apt(Ubuntu)

```bash
# apt install build-essential cmake golang-go libfltk1.4-dev libjsoncpp-dev libsdl3{,-image,-ttf}-dev
```

#### Build ninixw

```bash
$ git clone --recursive https://github.com/Tatakinov/ninixw
$ cd ninixw
$ ./build.sh
# cp -Pr ninixw /opt/ninixw
```

### Windows

Download zip from Release

or

```
> git clone --recursive https://github.com/Tatakinov/ninixw
> cd ninixw
> ./build.ps1
> cp win /path/to/ninixw
```

## LICENSE

ninixwはMIT Licenseです。

ninix-kagari側と似た処理があるかもしれませんが、そちらも私が書いたコードのはずなので問題ないはずです。
いや、ninix-ayaの時点からあったコードだぞ、という場合はご連絡ください。
