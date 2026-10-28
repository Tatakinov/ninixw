#ifndef NINIXW_SHM_SHM_H_
#define NINIXW_SHM_SHM_H_

#include <cstdint>
#include <string>

#if defined(_WIN32) || defined(WIN32)

#define WIN32_LEAN_AND_MEAN
#include <windows.h>

#else

#include <fcntl.h>
#include <semaphore.h>
#include <sys/mman.h>
#include <sys/stat.h>
#include <unistd.h>

#endif // WIN32

constexpr int NINIXW_CREAT = 0x1;

std::string readShm(int &err);

class Shm {
    private:
        bool valid_;
#if defined(_WIN32) || defined(WIN32)
        HANDLE mutex_;
        HANDLE fmo_;
#endif // WIN32
    public:
        Shm(const std::string &data, int flag);
        ~Shm();
        bool valid() const {
            return valid_;
        }
};

#endif // NINIXW_SHM_SHM_H_
