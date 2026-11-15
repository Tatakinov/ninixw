#include "contextmenu.h"

#include <iostream>

namespace {
    auto cb = [](Fl_Widget *w, void *data) {
        auto *v = static_cast<Json::Value *>(data);
        Json::StreamWriterBuilder builder;
        builder["indentation"] = "";
        std::cout << Json::writeString(builder, *v) << std::endl;
    };
}

ContextMenuWindow::ContextMenuWindow(const std::string &json_str, int x, int y) : Fl_Window(200, 200, "Context Menu"), once_(true), x_(x), y_(y) {
    shape_.resize(200 * 200 * 4);
    Fl_RGB_Image img(shape_.data(), 200, 200, 4);
    shape(img);

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
    if (e == FL_FOCUS && once_) {
        once_ = false;
        if (list_.size() > 0) {
            const Fl_Menu_Item *item = list_[0].popup(x_, y_);
            if (item != nullptr) {
                item->do_callback(this);
            }
        }
        hide();
        return 1;
    }
    return Fl_Window::handle(e);
}
