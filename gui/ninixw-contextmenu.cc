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
    if (argc < 3) {
        return 0;
    }
    int x = 0, y = 0;
    {
        std::istringstream iss(argv[1]);
        iss >> x;
    }
    {
        std::istringstream iss(argv[2]);
        iss >> y;
    }
    std::string line;
    std::getline(std::cin, line);
    if (line.ends_with("\r")) {
        line = line.substr(0, line.length() - 1);
    }
    Fl_Window::default_xclass("io.github.tatakinov.gui.contextmenu");
    auto *w = new ContextMenuWindow(line, x, y);
    //w->border(0);
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
