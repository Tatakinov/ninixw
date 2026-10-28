#!/bin/sh

BINDIR=ninixw/bin
LIBEXECDIR=ninixw/libexec
LIBDIR=ninixw/lib

mkdir -p ${LIBEXECDIR} ${LIBDIR}

cd core && go build -o ../${BINDIR}/ninixw.exe && cd -
if [ $? -ne 0 ]; then
    echo "failed to build core"
    exit 1
fi

cmake -S . -B build && cmake --build build
if [ $? -ne 0 ]; then
    echo "failed to build module"
    exit 1
fi
cp build/*/ninixw-* build/sorakado/sorakado_builtin.exe ${LIBEXECDIR}

echo "Copy ninixw directory -> /opt/ninixw and append /opt/ninixw/bin to PATH and run ninixw"
