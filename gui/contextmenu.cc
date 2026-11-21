#include "contextmenu.h"

#include <iostream>

namespace {
    auto cb = [](Fl_Widget *w, void *data) {
        auto *v = static_cast<Json::Value *>(data);
        Json::StreamWriterBuilder builder;
        builder["indentation"] = "";
        std::cout << Json::writeString(builder, *v) << std::endl;
    };
    auto show = [](void *data) {
        ContextMenuWindow *w = static_cast<ContextMenuWindow *>(data);
        Fl_Menu_Item *menu = w->menu();
        const Fl_Menu_Item *item = menu->popup(w->pos_x(), w->pos_y());
        if (item != nullptr) {
            item->do_callback(w);
        }
        w->hide();
    };
}

ContextMenuWindow::ContextMenuWindow(const std::string &json_str, int x, int y) : Fl_Window(200, 200, "Context Menu"), once_(true), pos_x_(x), pos_y_(y) {
    shape_.resize(1 * 1 * 4);
    Fl_RGB_Image img(shape_.data(), 1, 1, 4);
    //shape(img);

    //clear_border();
    //box(FL_NO_BOX);

    Json::Reader reader;
    reader.parse(json_str, value_);
    build(value_);
}

void ContextMenuWindow::build(Json::Value &value) {
    if (!value.isArray()) {
        return;
    }
    for (int i = 0; !value[i].isNull(); i++) {
        auto &v = value[i];
        auto type = v["type"].asString();
        if (type == "submenu") {
            list_.push_back({strdup(v["caption"].asString().c_str()), 0, 0, 0, FL_SUBMENU});
            build(v["list"]);
        }
        if (type == "item") {
            list_.push_back({strdup(v["caption"].asString().c_str()), 0, cb, &v});
        }
        if (type == "check") {
            // TODO stub
            list_.push_back({strdup(v["caption"].asString().c_str()), 0, cb, &v, FL_MENU_TOGGLE});
        }
    }
    list_.push_back({0});
}

ContextMenuWindow::~ContextMenuWindow() {
}

int ContextMenuWindow::handle(int e) {
    if (e == FL_SHOW && once_) {
        once_ = false;
        if (list_.size() > 0) {
#if defined(_WIN32) || defined(WIN32)
            if (0 <= pos_x_ && pos_x_ < w() && 0 <= pos_y_ && pos_y_ < h()) {
                shape_.resize(w() * h() * 4);
                shape_[(pos_y_ * w() + pos_x_) * 4 + 0] = 0xff;
                shape_[(pos_y_ * w() + pos_x_) * 4 + 1] = 0xff;
                shape_[(pos_y_ * w() + pos_x_) * 4 + 2] = 0xff;
                shape_[(pos_y_ * w() + pos_x_) * 4 + 3] = 0xff;
                Fl_RGB_Image img(shape_.data(), w(), h(), 4);
                //shape(img);
            }
#endif // WIN32
            Fl::add_timeout(1, show, this);
        }
        else {
            hide();
        }
        return 1;
    }
    return Fl_Window::handle(e);
}
