#include "contextmenu.h"

#include <iostream>

namespace {
    auto cb_show = [](void *data) {
        ContextMenuWindow *w = static_cast<ContextMenuWindow *>(data);
        Fl_Menu_Item *menu = w->menu();
        const Fl_Menu_Item *item = menu->popup(Fl::event_x(), Fl::event_y());
        if (item != nullptr) {
            item->do_callback(w);
        }
        w->hide();
    };
}

ContextMenuWindow::ContextMenuWindow(std::vector<Fl_Menu_Item> &list) : Fl_Window(200, 200, "Context Menu"), once_(true), list_(list) {
    shape_.resize(1 * 1 * 4);
    Fl_RGB_Image img(shape_.data(), 1, 1, 4);
    shape(img);

    clear_border();
    box(FL_NO_BOX);
}

ContextMenuWindow::~ContextMenuWindow() {
}

int ContextMenuWindow::handle(int e) {
    if (e == FL_SHOW && once_) {
        once_ = false;
        if (list_.size() > 0) {
#if defined(_WIN32) || defined(WIN32)
            int x = Fl::event_x();
            int y = Fl::event_y();
            if (0 <= x && x < w() && 0 <= y && y < h()) {
                shape_.resize(w() * h() * 4);
                shape_[(y * w() + x) * 4 + 0] = 0xff;
                shape_[(y * w() + x) * 4 + 1] = 0xff;
                shape_[(y * w() + x) * 4 + 2] = 0xff;
                shape_[(y * w() + x) * 4 + 3] = 0xff;
                Fl_RGB_Image img(shape_.data(), w(), h(), 4);
                shape(img);
            }
            else {
                shape_.resize(w() * h() * 4);
                shape_[0] = 0xff;
                shape_[1] = 0xff;
                shape_[2] = 0xff;
                shape_[3] = 0xff;
                Fl_RGB_Image img(shape_.data(), w(), h(), 4);
                shape(img);
            }
#endif // WIN32
            Fl::add_timeout(0.1, cb_show, this);
        }
        else {
            hide();
        }
        return 1;
    }
    return Fl_Window::handle(e);
}
