#include <iostream>
#include <memory>
#include <string>

#include "shm.h"

int main(int argc, char *argv[]) {
    if (argc < 2) {
        std::cerr << "too few argument" << std::endl;
        return 0;
    }
    {
        std::string arg = argv[1];
        if (arg == "-g") {
            int err = 0;
            auto path = readShm(err);
            std::cout << path << std::endl;
            return err;
        }
    }
    std::string path;
    int flag = 0;
    for (int i = 1; i < argc; i++) {
        std::string arg = argv[i];
        if (arg == "-n") {
            flag ^= NINIXW_CREAT;
        }
        else {
            path = arg;
        }
    }
    if (path.empty()) {
        std::cerr << "empty path" << std::endl;
        return 0;
    }
    auto shm = std::make_unique<Shm>(path, flag);
    if (!shm->valid()) {
        std::cerr << "shm failed" << std::endl;
        return 0;
    }
    std::cout.write("ok", 2);
    // closeされるまで待つ
    while (std::cin >> path);
    return 0;
}
