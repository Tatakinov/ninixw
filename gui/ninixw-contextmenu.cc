#include <iostream>
#include <sstream>
#include <string>
#include <vector>

#include <FL/Fl.H>
#include <FL/Fl_Double_Window.H>
#include <FL/Fl_Menu_Item.H>
#include <json/json.h>

#include "contextmenu.h"

int main() {
    std::string line;
    std::getline(std::cin, line);
    if (line.ends_with("\r")) {
        line = line.substr(0, line.length() - 1);
    }
    Fl_Window::default_xclass("io.github.tatakinov.gui.contextmenu");
    auto *w = new ContextMenuWindow(line);
    w->border(0);
    w->resizable(w);
    w->end();
    w->show();
    w->maximize();
    return Fl::run();
}
