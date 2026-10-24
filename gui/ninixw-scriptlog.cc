#include <iostream>
#include <sstream>
#include <string>
#include <thread>
#include <vector>

#include <FL/Fl.H>
#include <FL/Fl_Double_Window.H>
#include <FL/Fl_Table_Row.H>
#include <FL/fl_draw.H>

namespace {
    const std::vector<std::string> header = {
        "Time",
        "Ghost",
        "Status",
        "Script",
    };
}

struct Data {
    std::string time;
    std::string ghost;
    std::string status;
    std::string script;
};

class TableColumn : public Fl_Table_Row {
    private:
        std::vector<Data> data_;
    protected:
        void draw_cell(TableContext context, int r, int c, int x, int y, int w, int h) override {
            switch (context) {
                case CONTEXT_COL_HEADER:
                    fl_draw_box(FL_THIN_UP_BOX, x, y, w, h, FL_BACKGROUND_COLOR);
                    if (c < header.size()) {
                        fl_color(FL_BLACK);
                        fl_draw(header[c].c_str(), x, y, w, h, FL_ALIGN_CENTER);
                    }
                    break;
                case CONTEXT_CELL:
                    if (row_selected(r)) {
                        fl_draw_box(FL_FLAT_BOX, x, y, w, h, FL_BLUE);
                        fl_color(FL_WHITE);
                    }
                    else {
                        fl_draw_box(FL_FLAT_BOX, x, y, w, h, FL_WHITE);
                        fl_color(FL_BLACK);
                    }
                    if (r < data_.size()) {
                        switch (c) {
                            case 0:
                                fl_draw(data_[r].time.c_str(), x, y, w, h, FL_ALIGN_LEFT);
                                break;
                            case 1:
                                fl_draw(data_[r].ghost.c_str(), x, y, w, h, FL_ALIGN_LEFT);
                                break;
                            case 2:
                                fl_draw(data_[r].status.c_str(), x, y, w, h, FL_ALIGN_LEFT);
                                break;
                            case 3:
                                fl_draw(data_[r].script.c_str(), x, y, w, h, FL_ALIGN_LEFT);
                                break;
                            default:
                                // nop
                                break;
                        }
                    }
                default:
                    break;
            }
        }
    public:
        TableColumn(int x, int y, int w, int h) : Fl_Table_Row(x, y, w, h) {
            rows(0);
            cols(4);
            col_header(1);
            row_height_all(25);
            col_width_all(100);
            type(SELECT_SINGLE);
            end();
        }
        void append(const Data &data) {
            data_.insert(data_.begin(), data);
            while (data_.size() > 50) {
                data_.pop_back();
            }
            rows(data_.size());
            redraw();
        }
};

int main() {
    Fl_Double_Window w(400, 200, "Script Log");
    TableColumn t(20, 20, 360, 160);
    w.end();
    auto th = std::thread([&]() {
        std::string line;
        while (std::getline(std::cin, line)) {
            Data data;
            {
                std::string tmp;
                std::istringstream iss(line);
                char delim = '\x01';
                std::getline(iss, tmp, delim);
                data.time = tmp;
                std::getline(iss, tmp, delim);
                data.ghost = tmp;
                std::getline(iss, tmp, delim);
                data.status = tmp;
                std::getline(iss, tmp, '\0');
                data.script = tmp;
            }
            Fl::lock();
            t.append(data);
            Fl::unlock();
            Fl::awake();
        }
        Fl::lock();
        w.hide();
        Fl::unlock();
        Fl::awake();
    });
    w.show();
    Fl::run();
    th.join();
}
