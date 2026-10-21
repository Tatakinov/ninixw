#include "shm.h"

#include <cstdint>
#include <cstdlib>
#include <cstring>
#include <errno.h>
#include <fcntl.h>
#include <limits.h>
#include <semaphore.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>

struct shm_t {
    sem_t sem;
    uint32_t size;
    char buf[PATH_MAX];
};

std::string readShm(int &err) {
    int fd = shm_open("/ninix", O_RDWR, 0);
    if (fd == -1) {
        return "";
    }
    shm_t *mem = static_cast<shm_t *>(mmap(NULL, sizeof(*mem), PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0));
    if (mem == MAP_FAILED) {
        err = 1;
        return "";
    }
    if (sem_trywait(&mem->sem) == -1) {
        err = 1;
        return "";
    }
    std::string data(mem->buf, mem->size);
    return data;
}

Shm::Shm(const std::string &data, int flag) : valid_(false) {
    shm_t *mem = nullptr;
    if (flag & NINIXW_CREAT) {
        int err = shm_unlink("/ninix");
        if (err == -1) {
            return;
        }
    }
    int fd = shm_open("/ninix", O_CREAT | O_EXCL | O_RDWR, S_IRUSR | S_IWUSR);
    if (fd == -1) {
        return;
    }
    if (ftruncate(fd, sizeof(shm_t)) == -1) {
        goto error;
    }
    mem = static_cast<shm_t *>(mmap(NULL, sizeof(*mem), PROT_READ | PROT_WRITE, MAP_SHARED, fd, 0));
    if (mem == MAP_FAILED) {
        goto error;
    }
    if (sem_init(&mem->sem, 1, 0) == -1) {
        goto error;
    }
    mem->size = data.length();
    memcpy(mem->buf, data.data(), mem->size);
    if (sem_post(&mem->sem) == -1) {
        goto error;
    }
    valid_ = true;
    return;
error:
    shm_unlink("/ninix");
}

Shm::~Shm() {
    if (!valid_) {
        return;
    }
    shm_unlink("/ninix");
}
