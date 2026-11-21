#include <iostream>
#include <sstream>
#include <string>
#include <vector>

#include <FL/Fl.H>
#if FL_API_VERSION < 10400
#if defined(_WIN32) || defined(WIN32)
#include <FL/x.h>
#endif // WIN32
#endif // FL_API_VERSION < 10400

#include "contextmenu.h"

namespace {
    auto cb = [](Fl_Widget *w, void *data) {
        auto *v = static_cast<Json::Value *>(data);
        Json::StreamWriterBuilder builder;
        builder["indentation"] = "";
        std::cout << Json::writeString(builder, *v) << std::endl;
    };

    void build(Json::Value &value, std::vector<Fl_Menu_Item> &list) {
        if (!value.isArray()) {
            return;
        }
        for (int i = 0; !value[i].isNull(); i++) {
            auto &v = value[i];
            auto type = v["type"].asString();
            if (type == "submenu") {
                list.push_back({strdup(v["caption"].asString().c_str()), 0, 0, 0, FL_SUBMENU});
                build(v["list"], list);
            }
            if (type == "item") {
                list.push_back({strdup(v["caption"].asString().c_str()), 0, cb, &v});
            }
            if (type == "check") {
                // TODO stub
                list.push_back({strdup(v["caption"].asString().c_str()), 0, cb, &v, FL_MENU_TOGGLE});
            }
        }
        list.push_back({0});
    }
}

int main(int argc, char *argv[]) {
    std::string line;
    std::getline(std::cin, line);
    if (line.ends_with("\r")) {
        line = line.substr(0, line.length() - 1);
    }
    Json::Value value;
    Json::Reader reader;
    reader.parse(line, value);
    std::vector<Fl_Menu_Item> list;
    build(value, list);
    Fl_Window::default_xclass("io.github.tatakinov.gui.contextmenu");
    auto *w = new ContextMenuWindow(list);
    w->resizable(w);
    w->end();
    w->show();
#if FL_API_VERSION >= 10400
    w->maximize();
#else
#if defined(_WIN32) || defined(WIN32)
    ShowWindow(fl_xid(w), SW_MAXIMIZE);
#else
    // TODO stub
#endif // WIN32
#endif // FL_API_VERSION >= 10400
    return Fl::run();
}
