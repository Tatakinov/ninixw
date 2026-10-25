#include "lib.h"

#include <cassert>
#include <cstdlib>
#include <cstring>
#include <dlfcn.h>
#include <iostream>
#include <sstream>
#include <unordered_map>

namespace {
    std::string trimSpace(const std::string &str) {
        int pos = 0;
        for (int i = 0; i < str.length(); i++) {
            if (str[i] == ' ' || str[i] == '\t') {
                pos++;
            }
            else {
                break;
            }
        }
        int n = str.length() - pos;
        for (int i = str.length() - 1; i >= 0; i--) {
            if (str[i] == ' ' || str[i] == '\t') {
                n--;
            }
            else {
                break;
            }
        }
        return str.substr(pos, n);
    }

    inline ffi_type *getFfiType(std::string type) {
        std::unordered_map<std::string, ffi_type *> m = {
            {"int", &ffi_type_sint},
            {"uint", &ffi_type_uint},
            {"long", &ffi_type_slong},
            {"ulong", &ffi_type_ulong},
            {"void*", &ffi_type_pointer},
        };
        assert(m.contains(type));
        return m.at(type);
    }
}

Library::Library(int argc, char *argv[]) : valid_(false), loaded_(false), module_(nullptr) {
    if (argc < 5) {
        std::cerr << "too few arguments" << std::endl;
        return;
    }
    std::string lib = argv[0];
    std::string p = argv[1];
    std::u8string u8p(p.begin(), p.end());
    path_ = u8p;

    std::unordered_map<std::string, ffi_info> map;
    for (int i = 2; i < argc; i++) {
        std::istringstream iss1(argv[i]);
        ffi_info info;
        std::string tmp;
        std::getline(iss1, tmp, ' ');
        info.ret = trimSpace(tmp);
        std::string orig_name;
        std::getline(iss1, tmp, ':');
        orig_name = trimSpace(tmp);
        std::getline(iss1, tmp, '(');
        info.name = trimSpace(tmp);
        std::getline(iss1, tmp, ')');
        std::istringstream iss2(tmp);
        while (std::getline(iss2, tmp, ',')) {
            info.args.push_back(trimSpace(tmp));
        }
        map[orig_name] = info;
    }

    if (!(map.contains("load") && map.contains("request") && map.contains("unload"))) {
        std::cerr << "each of load/request/unload not found" << std::endl;
        return;
    }

    load_i_ = map.at("load");
    if (load_i_.args.size() != 2) {
        std::cerr << "load 2 <=> load " << load_i_.args.size() << std::endl;
        return;
    }
    request_i_ = map.at("request");
    if (request_i_.args.size() != 3) {
        std::cerr << "request 3 <=> request " << request_i_.args.size() << std::endl;
        return;
    }
    unload_i_ = map.at("unload");
    if (unload_i_.args.size() != 1) {
        std::cerr << "unload 1 <=> unload " << unload_i_.args.size() << std::endl;
        return;
    }

    module_ = dlopen(lib.c_str(), RTLD_LAZY);
    if (module_ == nullptr) {
        std::cerr << "failed to dlopen()" << std::endl;
        return;
    }
    load_f_ = dlsym(module_, load_i_.name.c_str());
    if (load_f_ == nullptr) {
        std::cerr << "failed to dlsym " << load_i_.name << std::endl;
        goto error;
    }
    request_f_ = dlsym(module_, request_i_.name.c_str());
    if (request_f_ == nullptr) {
        std::cerr << "failed to dlsym " << request_i_.name << std::endl;
        goto error;
    }
    unload_f_ = dlsym(module_, unload_i_.name.c_str());
    if (unload_f_ == nullptr) {
        std::cerr << "failed to dlsym " << unload_i_.name << std::endl;
        goto error;
    }
    valid_ = true;
    return;
error:
    dlclose(module_);
}

Library::~Library() {
    if (loaded_) {
        unload();
    }
    if (valid_) {
        dlclose(module_);
    }
}

bool Library::load() {
    if (!valid_) {
        std::cerr << "invalid library" << std::endl;
        return false;
    }
    ffi_cif cif;
    std::vector<ffi_type *> arg_types;
    arg_types.reserve(load_i_.args.size());
    std::vector<void *> arg_values;
    arg_values.reserve(load_i_.args.size());
    for (auto &t : load_i_.args) {
        arg_types.push_back(getFfiType(t));
    }
    ffi_status status = ffi_prep_cif(&cif, FFI_DEFAULT_ABI, load_i_.args.size(), getFfiType(load_i_.ret), arg_types.data());
    if (status != FFI_OK) {
        std::cerr << "invalid ffi status" << std::endl;
        return false;
    }
    std::string path = path_.string() + std::filesystem::path::preferred_separator;
    char *p = strdup(path.c_str());
    long len = path.length();
    arg_values.push_back(static_cast<void *>(&p));
    arg_values.push_back(static_cast<void *>(&len));
    ffi_call(&cif, FFI_FN(load_f_), &id_, arg_values.data());
    loaded_ = true;
    return true;
}

std::string Library::request(const std::string &req) {
    if (!valid_ || !loaded_) {
        std::cerr << "invalid library" << std::endl;
        return "";
    }
    ffi_cif cif;
    std::vector<ffi_type *> arg_types;
    arg_types.reserve(request_i_.args.size());
    std::vector<void *> arg_values;
    arg_values.reserve(request_i_.args.size());
    for (auto &t : request_i_.args) {
        arg_types.push_back(getFfiType(t));
    }
    ffi_status status = ffi_prep_cif(&cif, FFI_DEFAULT_ABI, request_i_.args.size(), getFfiType(request_i_.ret), arg_types.data());
    if (status != FFI_OK) {
        std::cerr << "invalid ffi status" << std::endl;
        return "";
    }
    char *r = strdup(req.c_str());
    long len = req.length();
    long *plen = &len;
    arg_values.push_back(static_cast<void *>(&id_));
    arg_values.push_back(static_cast<void *>(&r));
    arg_values.push_back(static_cast<void *>(&plen));
    ffi_sarg result;
    ffi_call(&cif, FFI_FN(request_f_), &result, arg_values.data());
    char *ret = reinterpret_cast<char *>(result);
    if (ret == nullptr || len == 0) {
        std::cerr << "invalid return value of request()" << std::endl;
        return "";
    }
    std::string res(ret, len);
    free(ret);
    return res;
}

bool Library::unload() {
    if (!valid_ || !loaded_) {
        std::cerr << "invalid library" << std::endl;
        return false;
    }
    ffi_cif cif;
    std::vector<ffi_type *> arg_types;
    arg_types.reserve(unload_i_.args.size());
    std::vector<void *> arg_values;
    arg_values.reserve(unload_i_.args.size());
    for (auto &t : unload_i_.args) {
        arg_types.push_back(getFfiType(t));
    }
    ffi_status status = ffi_prep_cif(&cif, FFI_DEFAULT_ABI, unload_i_.args.size(), getFfiType(unload_i_.ret), arg_types.data());
    if (status != FFI_OK) {
        std::cerr << "invalid ffi status" << std::endl;
        return false;
    }
    arg_values.push_back(static_cast<void *>(&id_));
    ffi_sarg result;
    ffi_call(&cif, FFI_FN(unload_f_), &result, arg_values.data());
    id_ = reinterpret_cast<long>(result);
    return true;
}
