#include "shm.h"

#include <cstdint>
#include <windows.h>

struct shm_t {
    uint32_t size;
    char buf[MAX_PATH];
};

std::string readShm() {
    shm_t *mem;
    HANDLE fmo;
    HANDLE mutex = OpenMutex(SYNCHRONIZE, FALSE, "/ninix_mutex");
    if (mutex == nullptr) {
        return "";
    }
    auto result = WaitForSingleObject(mutex, INFINITE);
    if (result != WAIT_OBJECT_0) {
        goto error_wait;
    }
    fmo = OpenFileMapping(FILE_MAP_ALL_ACCESS, FALSE, "/ninix");
    if (fmo == nullptr) {
        goto error_wait;
    }
    mem = static_cast<shm_t *>(MapViewOfFile(fmo_, FILE_MAP_ALL_ACCESS, 0, 0, 0));
    if (mem == nullptr) {
        goto error_map;
    }
    std::string data(mem->buf, mem->size());
    return data;
error_map:
    CloseHandle(mutex);
error_wait:
    CloseHandle(fmo);
    return "";
}

Shm::Shm(const std::string &data, int flag) {
    shm_t *mem;
    fmo_ = CreateFileMapping(INVALID_HANDLE_VALUE, NULL, PAGE_READWRITE, 0, sizeof(struct shm_t), "/ninix");
    if (fmo_ == nullptr) {
        return;
    }
    mem = static_cast<shm_t *>(MapViewOfFile(fmo_, FILE_MAP_ALL_ACCESS, 0, 0, 0));
    if (mem == nullptr) {
        goto error_map;
    }
    mutex_ = CreateMutex(NULL, TRUE, "/ninix_mutex");
    if (mutex_ == NULL) {
        goto error_mutex;
    }
    mem->size = data.length();
    memcpy(mem->buf, data.data(), data.length());
    ReleaseMutex(mutex_);
    valid_ = true;
    return;
error_mutex:
    UnmapViewOfFile(mem);
    if (0) {
error_map:
    CloseHandle(fmo_);
}

Shm::~Shm() {
    if (!valid_) {
        return;
    }
    CloseHandle(mutex_);
    CloseHandle(fmo);
}
