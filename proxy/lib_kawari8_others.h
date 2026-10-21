#ifndef NINIXW_PROXY_LIB_KAWARI8_OTHERS_H_
#define NINIXW_PROXY_LIB_KAWARI8_OTHERS_H_

#include "lib.h"

class Kawari8Library : public Library {
    private:
        unsigned int id_;
        func_t initialize_;
        func_t finalize_;
        func_t free_;
    public:
        Kawari8Library(int argc, char *argv[]);
        ~Kawari8Library();

        bool load() override;
        std::string request(const std::string &req) override;
        bool unload() override;
};

#endif // NINIXW_PROXY_LIB_KAWARI8_OTHERS_H_
