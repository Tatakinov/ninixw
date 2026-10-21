# ninixw

## これは何？

[ninix-kagari](https://github.com/Tatakinov/ninix-kagari)の精神的後継。

元プロジェクトはRuby + GTKだったのをGolang + SDLで新たに書いたもの。

## 名前の由来

Nin'i for X/Wayland/Windows

## Requirements

Required(Linux):

- go (core)
- cmake (gui, proxy, shm, sorakado)
- libffi (proxy)
- fltk1.3 (gui)
- SDL3 (sorakado)
- SDL3-image (sorakado)
- SDL3-ttf (sorakado)
- jsoncpp (sorakado)
- fontconfig (sorakado)
- wayland-client (sorakado)

Required(Windows):

- go (core)
- C++ compiler (gui, proxy, shm, sorakado, shiori)
- cmake (gui, proxy, shm, sorakado)
- libffi (proxy)
- fltk1.3 (gui)
- SDL3 (sorakado)
- SDL3-image (sorakado)
- SDL3-ttf (sorakado)
- jsoncpp (sorakado)
- openssl (shiori/aosora)

Optional(Linux/Windows):

- onnxruntime (sorakado)

## How to build

### Linux

```bash
$ git clone --recursive https://github.com/Tatakinov/ninixw
$ cd ninixw
$ ./build.sh
$ cp ninixw /opt/ninixw
```

### Windows

*stub*

## Note

Linuxでのみ動作確認しています。Windowsは多分動かない。

## LICENSE

ninixwはMIT Licenseです。

ninix-kagari側と似た処理があるかもしれませんが、そちらも私が書いたコードのはずなので問題ないはずです。
いや、ninix-ayaの時点からあったコードだぞ、という場合はご連絡ください。
