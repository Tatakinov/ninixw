#include <cstdint>
#include <iostream>
#include <memory>
#include <string>
#include <vector>

#if defined(_WIN32) || defined(WIN32)

#define WIN32_LEAN_AND_MEAN
#include <windows.h>

#else

#include "lib_pasta_others.h"
#include "lib_kawari8_others.h"

#endif // WIN32

#include "lib.h"

namespace {
    inline std::unique_ptr<Library> create(int argc, char *argv[]) {
#if defined(_WIN32) || defined(WIN32)
        return std::make_unique<Library>(argc, argv);
#else
        if (argc > 0 && std::string(argv[0]) == "pasta.dll") {
            return std::make_unique<PastaLibrary>(argc, argv);
        }
        else if (argc > 0 && std::string(argv[0]) == "shiori.dll") {
            return std::make_unique<Kawari8Library>(argc, argv);
        }
        else {
            return std::make_unique<Library>(argc, argv);
        }
#endif // WIN32
    }
}

int main(int argc, char *argv[]) {
    if (argc < 4) {
        return 1;
    }
    std::unique_ptr<Library> library;
    std::string command = argv[1];
    if (command != "--dry-run" && command != "--run") {
        return 1;
    }
    library = create(argc - 2, &argv[2]);
    if (command == "--dry-run") {
        if (library->valid()) {
            return 0;
        }
        else {
            return 2;
        }
    }
    if (!library->valid()) {
        return 2;
    }
#if defined(_WIN32) || defined(WIN32)
    _setmode(_fileno(stdin), _O_BINARY);
    _setmode(_fileno(stdout), _O_BINARY);
#endif // WIN32
    library->load();
    uint32_t len;
    while (true) {
        std::cin.read(reinterpret_cast<char *>(&len), sizeof(uint32_t));
        if (std::cin.eof() || len == 0) {
            break;
        }
        std::string request;
        request.resize(len);
        std::cin.read(request.data(), len);
        if (std::cin.gcount() < len) {
            return 3;
        }
        std::string response = library->request(request);
        len = response.size();
        std::cout.write(reinterpret_cast<char *>(&len), sizeof(uint32_t));
        std::cout.write(response.data(), len);
    }
    return 0;
}
