#ifndef NINIXW_PROXY_LIB_PASTA_OTHERS_H_
#define NINIXW_PROXY_LIB_PASTA_OTHERS_H_

#include <memory>
#include <optional>

#include "third_party/pasta/shiori4_abi.h"

#include "lib.h"

class PastaLibrary : public Library {
    private:
        std::optional<const ShioriApi *>api_;
        std::unique_ptr<ShioriHostApi> host_;
        uint32_t abi_version_;
        void *instance_;
    public:
        PastaLibrary(int argc, char *argv[]);
        ~PastaLibrary() {}

        bool load() override;
        std::string request(const std::string &req) override;
        bool unload() override;
};

#endif // NINIXW_PROXY_LIB_PASTA_OTHERS_H_
