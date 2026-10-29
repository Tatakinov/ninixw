$OnlyFontList = $Env:ONLY_FONTLIST ?? "OFF"

mkdir win/internal/gui
mkdir win/internal/proxy
mkdir win/internal/shm
mkdir win/internal/sorakado/sorakado_builtin

pushd core && go build -o ../win/ninixw.exe . && popd
cmake -S . -B build_i686 -DONLY_FONTLIST=$OnlyFontList -DBUILD_PROXY="ON" -A Win32 && cmake --build build --config Release
cmake -S . -B build_x86_64 -DONLY_FONTLIST=$OnlyFontList && cmake --build build --config Release
Copy-Item -Path build/gui/Release/* -Destination win/internal/gui -Recurse
Copy-Item -Path build/proxy/Release/* -Destination win/internal/proxy -Recurse
Copy-Item -Path build/shm/Release/* -Destination win/internal/shm -Recurse
Copy-Item -Path build/sorakado/Release/* -Destination win/internal/sorakado/sorakado_builtin -Recurse
