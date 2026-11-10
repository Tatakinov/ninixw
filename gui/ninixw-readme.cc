#include <iostream>
#include <sstream>
#include <string>

#include <FL/Fl.H>
#include <FL/Fl_Double_Window.H>
#include <FL/Fl_Button.H>
#include <FL/Fl_Text_Display.H>

#if defined(_WIN32) || defined(WIN32)
#define WIN32_LEAN_AND_MEAN
#include <windows.h>
#else
#include <iconv.h>
#endif // WIN32

void onOK(Fl_Widget *w, void *data) {
    static_cast<Fl_Double_Window *>(data)->hide();
}

#if defined(_WIN32) || defined(WIN32)
std::string toUTF8(std::string in) {
    int w_size = MultiByteToWideChar(932, MB_PRECOMPOSED, in.data(), in.length(), nullptr, 0);
    if (w_size == 0) {
        return in;
    }
    std::wstring w_tmp;
    w_tmp.resize(w_size);
    MultiByteToWidtChar(932, MB_PRECOMPOSED, in.data(), in.length(), w_tmp.data(), w_tmp.length());
    int out_size = WideCharToMultiByte(CP_UTF8, WC_ERR_INVALID_CHARS, w_tmp.data(), w_tmp.length(), nullptr, 0, nullptr, nullptr);
    if (out_size == 0) {
        return in;
    }
    std::string out;
    out.resize(out_size);
    WideCharToMultiByte(CP_UTF8, WC_ERR_INVALID_CHARS, w_tmp.data(), w_tmp.length(), out.data(), out.length(), nullptr, nullptr);
    return out;
}
#else
std::string toUTF8(std::string in) {
    auto conv = iconv_open("UTF8", "CP932");
    if (conv == (iconv_t)-1) {
        return in;
    }
    char *in_p = in.data();
    size_t in_l = in.length();
    std::string out;
    out.resize(in.length() * 2); // 2倍あれば十分
    char *out_p = out.data();
    size_t out_l = out.length();
    while (true) {
        if (iconv(conv, &in_p, &in_l, &out_p, &out_l) == (size_t)-1) {
            return in;
        }
        if (in_l == 0) {
            break;
        }
    }
    return out.substr(0, out.length() - out_l);
}
#endif // WIN32

int main(int argc, char *argv[]) {
    if (argc < 2) {
        return 1;
    }
    std::ostringstream oss;
    std::string line;
    while (std::getline(std::cin, line)) {
        oss << line << "\n";
    }
    std::string in = oss.str();
    if (std::string(argv[1]) == "Shift_JIS") {
        in = toUTF8(in);
    }
    Fl_Window::default_xclass("io.github.tatakinov.gui.readme");
    auto *w = new Fl_Double_Window(480, 320);
    {
        auto *o = new Fl_Button(401, 283, 64, 22, "OK");
        o->callback((Fl_Callback*)onOK, w);
    }
    {
        auto *o = new Fl_Text_Display(15, 15, 450, 260);
        auto *b = new Fl_Text_Buffer();
        b->append(in.c_str());
        o->buffer(b);
    }
    w->end();
    w->show();
    return Fl::run();
}
