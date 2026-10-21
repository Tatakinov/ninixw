#include <cstdlib>

#include "msgbox.h"

namespace {
    int ret = 1;
}

void onOK(Fl_Widget *w, void *data) {
    ret = 0;
    static_cast<Fl_Double_Window *>(data)->hide();
}

void onCancel(Fl_Widget *w, void *data) {
    ret = 1;
    static_cast<Fl_Double_Window *>(data)->hide();
}

int main(int argc, char *argv[]) {
    if (argc < 4) {
        return 2;
    }
    auto *w = make_window(argv[1], argv[2], argv[3]);
    w->show();
    Fl::run();
    return ret;
}
