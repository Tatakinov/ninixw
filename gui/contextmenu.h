#ifndef NINIXW_GUI_CONTEXTMENU_H_
#define NINIXW_GUI_CONTEXTMENU_H_

#include <string>
#include <vector>

#include <FL/Fl.H>
#include <FL/Fl_Window.H>
#include <FL/Fl_Menu_Item.H>
#include <json/json.h>

class ContextMenuWindow : public Fl_Window {
    private:
        bool once_;
        std::vector<uchar> shape_;
        std::vector<Fl_Menu_Item> list_;
        Json::Value value_;
    public:
        ContextMenuWindow(const std::string &json_str);
        ~ContextMenuWindow();

        void build(Json::Value &value);
        int handle(int e) override;
};

#endif // NINIXW_GUI_CONTEXTMENU_H_
