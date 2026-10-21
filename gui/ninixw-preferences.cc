#include <iostream>

#include "preferences.h"

void onButtonClick(Fl_Widget *w, void *data) {
    std::cout << "ok!" << std::endl;
}

int main() {
    auto *w = make_window();
    w->show();
    return Fl::run();
}
