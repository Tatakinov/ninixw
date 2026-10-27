#include "preferences.h"

#include <json/json.h>

#include "fontlist.hpp"

namespace {
    std::string get(const Json::Value &v, const char *k, std::string fallback = "") {
        if (v.isMember(k) && v[k].isString()) {
            return v[k].asString();
        }
        return fallback;
    }
}

Preferences::Preferences(const std::string &json_str) : original_(json_str) {
    Json::Reader reader;
    Json::Value value;
    reader.parse(json_str, value);
    auto family = fontlist::get_default_font();
    font_name = get(value, "font_name", family.name);
    font_size = get(value, "font_size", "12");
    shell_scale = get(value, "shell_scale", "100");
    default_balloon = get(value, "default_balloon", "");
    if (value.isMember("balloons")) {
        for (auto &v : value["balloons"]) {
            if (v.isString()) {
                balloon_list_.push_back(v.asString());
            }
        }
    }
}

Preferences::~Preferences() {}

std::string Preferences::serialize() const {
    Json::Value value;
    value["font_name"] = font_name;
    value["font_size"] = font_size;
    value["shell_scale"] = shell_scale;
    value["default_balloon"] = default_balloon;
    Json::StreamWriterBuilder builder;
    builder["indentation"] = "";
    return Json::writeString(builder, value);
}

std::string Preferences::original() const {
    return original_;
}

std::vector<std::string> Preferences::balloonList() const {
    return balloon_list_;
}
