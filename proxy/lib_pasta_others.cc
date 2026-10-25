#include "lib_pasta_others.h"

#include <cstdlib>
#include <cstring>
#include <dlfcn.h>

namespace {
    char failed;

    void release(ShioriBytes *p) {
        free(p->owner);
        *p = SHIORI_BYTES_EMPTY;
        return;
    }

    void allocate(void *ctx, const uint8_t *ptr, size_t len, ShioriBytes *out) {
        uint8_t *p = reinterpret_cast<uint8_t *>(malloc(len ? len : 1));
        if (p == nullptr) {
            *out = { nullptr, 0, &failed, nullptr };
            return;
        }
        if (len > 0) {
            memcpy(p, ptr, len);
        }
        *out = { p, len, p, release };
        return;
    };
}

PastaLibrary::PastaLibrary(int argc, char *argv[]) : Library(argc, argv), abi_version_(SHIORI_ABI_VERSION), instance_(nullptr) {
    if (argc < 2) {
        return;
    }
    std::string lib = argv[0];
    std::string p = argv[1];
    std::u8string u8p(p.begin(), p.end());
    path_ = u8p;

    module_ = dlopen(lib.c_str(), RTLD_LAZY);
    if (module_ == nullptr) {
        return;
    }
    auto get_api = reinterpret_cast<ShioriGetApiFn>(dlsym(module_, SHIORI_GET_API_SYMBOL));
    if (get_api == nullptr) {
        goto error;
    }
    api_ = get_api(abi_version_);
    if ((*api_) == nullptr) {
        abi_version_ = SHIORI_ABI_VERSION_3;
        api_ = get_api(abi_version_);
    }
    else {
        // FIXME implement v4-request
        return;
    }
    if (!api_ || !(*api_) || (*api_)->abi_version != abi_version_ || !SHIORI_API_IS_VALID((*api_)) || !SHIORI_HAS_FIELD((*api_), ShioriApi, notify)) {
        goto error;
    }
    switch (abi_version_) {
        case SHIORI_ABI_VERSION:
            host_ = std::make_unique<ShioriHostApi>(
                    sizeof(ShioriHostApi),
                    abi_version_,
                    nullptr,
                    [](void *ctx, const uint8_t *req, size_t len, ShioriBytes *out) {
                        return SHIORI_400_BAD_REQUEST;
                    },
                    [](void *ctx, const uint8_t *msg, size_t len) {
                        return SHIORI_204_NO_CONTENT;
                    },
                    [](void *ctx, int32_t level, const uint8_t *msg, size_t len) {
                        // nop
                    },
                    allocate
                    );
            break;
        case SHIORI_ABI_VERSION_3:
            host_ = std::make_unique<ShioriHostApi>(
                    sizeof(ShioriHostApi),
                    abi_version_,
                    nullptr,
                    nullptr,
                    nullptr,
                    nullptr,
                    allocate
                    );
            break;
        default:
            goto error;
    }
    valid_ = true;
    return;
error:
    dlclose(module_);
    return;
}

bool PastaLibrary::load() {
#if 0
    if (!valid_) {
        return false;
    }
    auto u8p = path_.u8string();
    std::string p(u8p.begin(), u8p.end());
    instance_ = (*api_)->load(host_.get(), reinterpret_cast<const uint8_t *>(p.data()), p.length());
    if (instance_ == nullptr) {
        dlclose(module_);
        valid_ = false;
        return false;
    }
    loaded_ = true;
#endif
    return true;
}

std::string PastaLibrary::request(const std::string &req) {
#if 0
    if (abi_version_ == SHIORI_ABI_VERSION) {
        // TODO stub
    }
    else if (abi_version_ == SHIORI_ABI_VERSION_3) {
        ShioriBytes out = SHIORI_BYTES_EMPTY;
        auto status = (*api_)->request(instance_, reinterpret_cast<const uint8_t *>(req.data()), req.length(), &out);
        bool copied = out.owner != &failed;
        if (copied) {
            // TODO error
        }
        std::string ret = "";
        if (out.len > 0) {
            ret = std::string(reinterpret_cast<const char *>(out.ptr), out.len);
        }
        shiori_bytes_release(&out);
        if (SHIORI_IS_SUCCESS(status) && copied) {
            return ret;
        }
        return "";
    }
#endif
    return "";
}

bool PastaLibrary::unload() {
    if (loaded_) {
        (*api_)->unload(instance_);
        loaded_ = false;
    }
    return true;
}
