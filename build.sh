#!/bin/sh

BINDIR=ninixw/bin
LIBEXECDIR=ninixw/libexec
LIBDIR=ninixw/lib

mkdir -p ${LIBEXECDIR} ${LIBDIR}

cd core && go build . && cp ninixw-core ../${BINDIR} && cd -
if [ $? -ne 0 ]; then
    echo "failed to build core"
    exit 1
fi
cd gui && cmake -S . -B build && cmake --build build && cp build/ninixw-* ../${LIBEXECDIR} && cd -
if [ $? -ne 0 ]; then
    echo "failed to build gui"
    exit 1
fi
cd proxy && cmake -S . -B build && cmake --build build && cp build/ninixw-proxy.exe ../${LIBEXECDIR} && cd -
if [ $? -ne 0 ]; then
    echo "failed to build proxy"
    exit 1
fi
cd shm && cmake -S . -B build && cmake --build build && cp build/ninixw-shm.exe ../${LIBEXECDIR} && cd -
if [ $? -ne 0 ]; then
    echo "failed to build shm"
    exit 1
fi
cd sorakado && cmake -S . -B build && cmake --build build && cp build/sorakado_builtin.exe ../${LIBEXECDIR} && cd -
if [ $? -ne 0 ]; then
    echo "failed to build sorakado"
    exit 1
fi

echo "Copy ninixw directory -> /opt/ninixw and append /opt/ninixw/bin to PATH and run ninixw"
