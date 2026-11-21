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

int main(int argc, char *argv[]) {
    std::string line;
    std::getline(std::cin, line);
    if (line.ends_with("\r")) {
        line = line.substr(0, line.length() - 1);
    }
    Json::Value value;
    Json::Reader reader;
    reader.parse(line, value);
    Fl_Window::default_xclass("io.github.tatakinov.gui.contextmenu");
    auto *w = new ContextMenuWindow(value);
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
