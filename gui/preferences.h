#ifndef NINIXW_GUI_PREFERENCES_H_
#define NINIXW_GUI_PREFERENCES_H_

#include <string>
#include <vector>

struct Preferences {
    private:
        std::string original_;
        std::vector<std::string> balloon_list_;
    public:
        std::string font_name;
        std::string font_size;
        std::string shell_scale;
        std::string default_balloon;

        Preferences(const std::string &json_str);
        ~Preferences();

        std::string serialize() const;
        std::string original() const;
        std::vector<std::string> balloonList() const;
};

#endif // NINIXW_GUI_PREFERENCES_H_
