#ifndef NINIXW_GUI_PREFERENCES_H_
#define NINIXW_GUI_PREFERENCES_H_

#include <string>

struct Preferences {
    private:
        std::string original_;
    public:
        std::string font_name;
        int font_size;
        int shell_scale;
        std::string default_balloon;

        Preferences(const std::string &json_str);
        ~Preferences();

        std::string serialize() const;

        std::string original() const;
};

#endif // NINIXW_GUI_PREFERENCES_H_
