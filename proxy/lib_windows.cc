#include "lib.h"

#define WIN32_LEAN_AND_MEAN
#include <windows.h>

using load_t = BOOL (__cdecl *)(HGLOBAL h, long len);
using unloat_t = BOOL (__cdecl *)();
using request_t = HGLOBAL (__cdecl *)(HGLOBAL h, long *len);

Library::Library(int argc, char *argv[]) : valid_(false), loaded_(false), module_(nullptr) {
    if (argc < 1) {
        return;
    }
    std::string lib = argv[0];
    std::u8string u8lib(lib.begin(), lib.end());
    std::filesystem::path plib = u8lib;
    std::string p = argv[1];
    std::u8string u8p(p.begin(), p.end());
    path_ = u8p;
    module_ = LoadLibraryW(plib.wstring().c_str());
    if (module_ == nullptr) {
        return;
    }
    load_f_ = GetProcAddress(module_, "load");
    loadu_f_ = GetProcAddress(module_, "loadu");
    request_f_ = GetProcAddress(module_, "request");
    unload_f_ = GetProcAddress(module_, "unload");
    if (load_f_ == nullptr || request_f_ == nullptr || unload_f_ == nullptr) {
        FreeLibrary(module_);
        return;
    }
    valid_ = true;
}

Library::~Library() {
    if (loaded_) {
        unload();
    }
    if (valid_) {
        FreeLibrary(module_);
    }
}

bool Library::load() {
    if (!valid_) {
        std::cerr << "invalid library" << std::endl;
        return false;
    }
    if (loadu_f_ == nullptr) {
        std::string p = path_.string();
        HGLOBAL h = GlobalAlloc(GPTR, p.length() + 1);
        if (h == nullptr) {
            FreeLibrary(module_);
            valid = false;
            return false;
        }
        memcpy(h, p.data(), p.length());
        auto f = reinterpret_cast<load_t>(load_f_);
        if (!f(h, p.length())) {
            FreeLibrary(module_);
            valid_ = false;
            return false;
        }
    }
    else {
        std::u8string u8p = path_.u8string();
        std::string p(u8p.begin(), u8p.end());
        HGLOBAL h = GlobalAlloc(GPTR, p.length() + 1);
        if (h == nullptr) {
            FreeLibrary(module_);
            valid_ = false;
            return false;
        }
        memcpy(h, p.c_str(), p.length());
        auto f = reinterpret_cast<load_t>(loadu_f_);
        if (!loadu(h, p.length())) {
            FreeLibrary(module_);
            valid_ = false;
            return 1;
        }
    }
    loaded_ = true;
}

std::string Library::request(const std::string &req) {
    if (!valid_ || !loaded_) {
        std::cerr << "invalid library" << std::endl;
        return "";
    }
    auto len = req.length();
    HGLOBAL h = GlobalAlloc(GPTR, len + 1);
    if (h == nullptr) {
        FreeLibrary(module_);
        valid_ = false;
        return "";
    }
    memcpy(h, req.data(), len);
    auto f = reinterpret_cast<request_t>(request_f_);
    h = f(h, &len);
    if (h == nullptr) {
        return "";
    }
    std::string response(reinterpret_cast<char *>(h), len);
    GlobalFree(h);
    return response;
}

bool Library::unload() {
    if (!valid_ || !loaded_) {
        std::cerr << "invalid library" << std::endl;
        return false;
    }
    auto f = reinterpret_cast<unload_t>(unload_f_);
    return f();
}
