#include <string>
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

namespace {
    std::vector<Data> data;
}

class TableColumn : public Fl_Table_Row {
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
                    if (r < data.size()) {
                        switch (c) {
                            case 0:
                                fl_draw(data[r].time.c_str(), x, y, w, h, FL_ALIGN_LEFT);
                                break;
                            case 1:
                                fl_draw(data[r].ghost.c_str(), x, y, w, h, FL_ALIGN_LEFT);
                                break;
                            case 2:
                                fl_draw(data[r].status.c_str(), x, y, w, h, FL_ALIGN_LEFT);
                                break;
                            case 3:
                                fl_draw(data[r].script.c_str(), x, y, w, h, FL_ALIGN_LEFT);
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
            rows(50);
            cols(4);
            col_header(1);
            row_height_all(25);
            col_width_all(100);
            type(SELECT_SINGLE);
            end();
        }
};

int main() {
    data = {
        {"01", "G1", "", "\\0test\\e"},
        {"02", "G2", "embed", "\\0test2aaaaaaaaaaaaaaaaaaaaa\\e"},
    };
    Fl_Double_Window w(400, 200, "Script Log");
    TableColumn t(20, 20, 360, 160);
    w.end();
    w.show();
    return Fl::run();
}
