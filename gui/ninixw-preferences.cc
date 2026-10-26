#include <iostream>
#include <memory>
#include <string>

#include <FL/Fl.H>
#include <FL/Fl_Double_Window.H>
#include <FL/Fl_Button.H>
#include <FL/Fl_Choice.H>
#include <FL/Fl_Input.H>
#include <json/json.h>

#include "fontlist.hpp"

#include "preferences.h"

namespace {
    std::unique_ptr<Preferences> prefs;
    auto font_list = fontlist::enumerate_font();
    std::vector<int> font_size_list = {
        8, 9,
        10, 11, 12, 14, 16, 18,
        20, 22, 24, 28, 32, 36,
        40, 44, 48, 56, 64, 72,
        80, 88, 96,
    };
    std::vector<int> shell_scale_list = {
        25, 40, 50, 80, 100, 125, 150, 175, 200, 250, 300, 350, 400,
    };

    void onOK(Fl_Widget *w, void *data) {
        std::cout << prefs->serialize() << std::endl;
    }
    void onApply(Fl_Widget *w, void *data) {
        std::cout << prefs->serialize() << std::endl;
    }
    void onCancel(Fl_Widget *w, void *data) {
        std::cout << prefs->original() << std::endl;
    }
    void onSelectFontName(Fl_Widget *w, void *data) {
        int index = reinterpret_cast<Fl_Choice *>(w)->value();
        prefs->font_name = font_list[index].name;
    }
    void onSelectFontSize(Fl_Widget *w, void *data) {
        int index = reinterpret_cast<Fl_Choice *>(w)->value();
        prefs->font_size = font_size_list[index];
    }
    void onSelectShellScale(Fl_Widget *w, void *data) {
        int index = reinterpret_cast<Fl_Choice *>(w)->value();
        prefs->shell_scale = shell_scale_list[index];
    }
    void onSelectBalloon(Fl_Widget *w, void *data) {
        int index = reinterpret_cast<Fl_Choice *>(w)->value();
        prefs->default_balloon = "balloon_terminal";
    }

}

int main() {
    std::string line;
    std::getline(std::cin, line);
    if (line.ends_with("\r")) {
        line = line.substr(0, line.length() - 1);
    }
    prefs = std::make_unique<Preferences>(line);
    auto *w = new Fl_Double_Window(480, 320);
    {
        auto *o = new Fl_Button(273, 283, 64, 22, "OK");
        o->callback((Fl_Callback*)onOK);
    } // Fl_Button* o
    {
        auto *o = new Fl_Button(337, 283, 64, 22, "Apply");
        o->callback((Fl_Callback*)onApply);
    } // Fl_Button* o
    {
        auto *o = new Fl_Button(401, 283, 64, 22, "Cancel");
        o->callback((Fl_Callback*)onCancel);
    } // Fl_Button* o
    {
        auto *o = new Fl_Choice(172, 15, 204, 20, "font_name");
        o->down_box(FL_BORDER_BOX);
        auto list = font_list;
        int index = -1;
        for (int i = 0; i < list.size(); i++) {
            o->add(list[i].name.c_str());
            if (list[i].name == prefs->font_name) {
                index = i;
            }
        }
        if (index == -1) {
            index = 0;
        }
        o->value(index);
        o->callback(onSelectFontName);
    } // Fl_Choice* o
    {
        auto* o = new Fl_Choice(172, 43, 204, 20, "font_size");
        o->down_box(FL_BORDER_BOX);
        auto list = font_size_list;
        int index = -1;
        for (int i = 0; i < list.size(); i++) {
            std::ostringstream oss;
            oss << list[i];
            auto s = oss.str();
            o->add(s.c_str());
            if (list[i] == prefs->font_size) {
                index = i;
            }
        }
        if (index == -1) {
            index = 0;
        }
        o->value(index);
        o->callback(onSelectFontSize);
    } // Fl_Choice* o
    {
        auto *o = new Fl_Choice(172, 71, 204, 20, "shell_scale");
        auto list = shell_scale_list;
        int index = -1;
        for (int i = 0; i < list.size(); i++) {
            std::ostringstream oss;
            oss << list[i];
            auto s = oss.str();
            o->add(s.c_str());
            if (list[i] == prefs->shell_scale) {
                index = i;
            }
        }
        if (index == -1) {
            index = 0;
        }
        o->value(index);
        o->callback(onSelectShellScale);
    } // Fl_Choice* o
    { Fl_Choice* o = new Fl_Choice(172, 99, 204, 20, "default_balloon");
      o->down_box(FL_BORDER_BOX);
      o->add("balloon_terminal");
      o->value(0);
      o->callback(onSelectBalloon);
    } // Fl_Choice* o
    w->end();
    w->show();
    return Fl::run();
}
