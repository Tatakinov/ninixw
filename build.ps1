$OnlyFontList = $Env:ONLY_FONTLIST ?? "OFF"

mkdir ninixw/internal/gui
mkdir ninixw/internal/proxy
mkdir ninixw/internal/shm
mkdir ninixw/internal/sorakado/sorakado_builtin
mkdir ninixw/internal/locale

pushd core && go build -o ../ninixw/ninixw.exe . && popd
cmake -S . -B build_i686 -DONLY_FONTLIST=$OnlyFontList -DBUILD_PROXY="ON" -A Win32 && cmake --build build_i686 --config Release
cmake -S . -B build_x86_64 -DONLY_FONTLIST=$OnlyFontList && cmake --build build_x86_64 --config Release
Copy-Item -Path build_x86_64/gui/Release/* -Destination ninixw/internal/gui -Recurse
Copy-Item -Path build_i686/proxy/Release/* -Destination ninixw/internal/proxy -Recurse
Copy-Item -Path build_x86_64/shm/Release/* -Destination ninixw/internal/shm -Recurse
Copy-Item -Path build_x86_64/sorakado/Release/* -Destination ninixw/internal/sorakado/sorakado_builtin -Recurse
Copy-Item -Path locale/* -Destination ninixw/internal/locale/*
