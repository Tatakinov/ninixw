#include <iostream>
#include <sstream>
#include <string>
#include <vector>

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
    w->border(0);
    w->resizable(w);
    w->end();
    w->show();
    w->maximize();
    return Fl::run();
}
