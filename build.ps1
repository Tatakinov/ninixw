mkdir -p win/internal/gui win/internal/proxy win/internal/shm win/internal/sorakado/sorakado_builtin

pushd core && go build -o ../win/ninixw.exe . && popd
cmake -S . -B build && cmake --build build --config Release
cp build/gui/Release/ninixw-* win/internal/gui/
cp build/proxy/Release/ninixw-* win/internal/proxy/
cp build/shm/Release/ninixw-* win/internal/shm/
cp build/sorakado/Release/sorakado_builtin.exe win/internal/sorakado/sorakado_builtin/
