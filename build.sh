#!/bin/sh

BINDIR=ninixw/bin
LIBEXECDIR=ninixw/libexec
LIBDIR=ninixw/lib
LOCALEDIR=ninixw/locale

if [ "$1" = "clean" ]; then
    rm -rf ./ninixw
    rm -rf ./build
    cd shiori/kawari/build/src && make -f gcc.mak clean && git checkout gcc.mak && cd -
    cd shiori/satori/satoriya/satori && make -f makefile.linux clean && cd -
    cd shiori/yaya && make -f makefile.linux clean && cd -
    exit 0
fi

# mkdir output directory
mkdir -p ${BINDIR} ${LIBEXECDIR} ${LIBDIR} ${LOCALEDIR}

# build core
cd core && go build -o ../${BINDIR}/ninixw-core && cd -
if [ $? -ne 0 ]; then
    echo "failed to build core"
    exit 1
fi

# build gui, proxy, shm, sorakado
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release && cmake --build build $@
if [ $? -ne 0 ]; then
    echo "failed to build module"
    exit 1
fi
cp build/*/ninixw-* build/sorakado/sorakado_builtin.exe ${LIBEXECDIR}/

# build kawari
cd shiori/kawari/build/src && sed -i -e 's/^MACH_TYPE = mingw/MACH_TYPE = linux/g' gcc.mak && make -f gcc.mak $@ && cd -
if [ $? -ne 0 ]; then
    echo "failed to build kawari"
fi
cp shiori/kawari/build/mach/linux/libshiori.so ${LIBDIR}/
ln -s libshiori.so ${LIBDIR}/shiori.dll
# build satori
cd shiori/satori/satoriya/satori && make -f makefile.linux $@ && cd -
if [ $? -ne 0 ]; then
    echo "failed to build satori"
fi
cp shiori/satori/satoriya/satori/libsatori.so ${LIBDIR}/
ln -s libsatori.so ${LIBDIR}/satori.dll
# build yaya
cd shiori/yaya && make -f makefile.linux $@ && cd -
if [ $? -ne 0 ]; then
    echo "failed to build yaya"
fi
cp shiori/yaya/libaya5.so ${LIBDIR}/
ln -s libaya5.so ${LIBDIR}/yaya.dll

cp -r locale/* ${LOCALEDIR}/

cp linux/bin/ninixw ${BINDIR}

# build complete
echo ""
echo ""
echo '`# cp -Pr ./ninixw /opt/ninixw` and append `/opt/ninixw/bin` to PATH and run `$ ninixw`'
