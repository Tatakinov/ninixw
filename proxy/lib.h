#ifndef NINIXW_PROXY_LIB_H_
#define NINIXW_PROXY_LIB_H_

#include <filesystem>

#if defined(_WIN32) || defined(WIN32)
#define WIN32_LEAN_AND_MEAN

#include <windows.h>

using module_t = HMODULE;
using func_t = FARPROC;

#else

#include <string>
#include <vector>
#include <ffi.h>

using module_t = void *;
using func_t = void *;

struct ffi_info {
    std::string ret;
    std::string name;
    std::vector<std::string> args;
};

#endif // WIN32

class Library {
    protected:
        bool valid_;
        bool loaded_;
        std::filesystem::path path_;
        module_t module_;
        func_t load_f_;
#if defined(_WIN32) || defined(WIN32)
        func_t loadu_f_;
#endif // WIN32
        func_t request_f_;
        func_t unload_f_;
#if !defined(_WIN32) && !defined(WIN32)
        ffi_sarg id_;
        ffi_info load_i_;
        ffi_info request_i_;
        ffi_info unload_i_;
#endif // !WIN32
    public:
        Library(int argc, char *argv[]);
        virtual ~Library();

        virtual bool load();
        virtual std::string request(const std::string &req);
        virtual bool unload();

        bool valid() const { return valid_; }
};

#endif // NINIXW_PROXY_LIB_H_
