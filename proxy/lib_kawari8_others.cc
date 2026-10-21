#include "lib_kawari8_others.h"

#include <cassert>
#include <dlfcn.h>

using initialize_t = int (*)();
using load_t = unsigned int (*)(const char *, long);
using request_t = const char *(*)(unsigned int, const char *, long *);
using unload_t = int (*)(unsigned int);
using free_t = void (*)(unsigned int, const char *);
using finalize_t = int (*)();

Kawari8Library::Kawari8Library(int argc, char *argv[]) : Library(argc, argv) {
    if (argc < 2) {
        return;
    }
    std::string lib = argv[0];
    std::string p = argv[1];
    std::u8string u8p(p.begin(), p.end());
    path_ = u8p;

    initialize_t f;

    module_ = dlopen(lib.c_str(), RTLD_LAZY);
    if (module_ == nullptr) {
        return;
    }
    initialize_ = dlsym(module_, "so_library_init");
    if (initialize_ == nullptr) {
        goto error;
    }
    load_f_ = dlsym(module_, "so_create");
    if (load_f_ == nullptr) {
        goto error;
    }
    request_f_ = dlsym(module_, "so_request");
    if (request_f_ == nullptr) {
        goto error;
    }
    unload_f_ = dlsym(module_, "so_dispose");
    if (unload_f_ == nullptr) {
        goto error;
    }
    free_ = dlsym(module_, "so_free");
    if (free_ == nullptr) {
        goto error;
    }
    finalize_ = dlsym(module_, "so_library_cleanup");
    if (finalize_ == nullptr) {
        goto error;
    }
    valid_ = true;
    f = reinterpret_cast<initialize_t>(initialize_);
    f();
    return;
error:
    dlclose(module_);
}

Kawari8Library::~Kawari8Library() {
    if (valid_) {
        finalize_t f = reinterpret_cast<finalize_t>(finalize_);
        f();
    }
}

bool Kawari8Library::load() {
    if (!valid_) {
        return false;
    }
    std::string path = path_.string();
    load_t f = reinterpret_cast<load_t>(load_f_);
    id_ = f(path.data(), path.length());
    loaded_ = true;
    return true;
}

std::string Kawari8Library::request(const std::string &req) {
    if (!valid_) {
        return "";
    }
    request_t req_f = reinterpret_cast<request_t>(request_f_);
    long len = req.length();
    const char *ret = req_f(id_, req.data(), &len);
    if (ret == nullptr || len == 0) {
        return "";
    }
    std::string res(ret, len);
    free_t free_f = reinterpret_cast<free_t>(free_);
    free_f(id_, ret);
    return res;
}

bool Kawari8Library::unload() {
    if (!valid_ || !loaded_) {
        return false;
    }
    unload_t f = reinterpret_cast<unload_t>(unload_f_);
    f(id_);
    return true;
}
