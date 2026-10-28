cmake_minimum_required(VERSION 3.30)
project(ninixw-gui CXX)

find_package(FLTK REQUIRED)
find_package(jsoncpp REQUIRED)

SET(INCLUDE_DIRS ${PROJECT_SOURCE_DIR} ${PROJECT_SOURCE_DIR}/libfontlist/include ${FLTK_INCLUDE_DIR} ${FLTK_INCLUDE_DIRS})

add_executable(ninixw-preferences ninixw-preferences.cc preferences.cc)
set_target_properties(ninixw-preferences PROPERTIES SUFFIX ".exe")
target_compile_options(ninixw-preferences PUBLIC -O2 -Wall)
target_compile_features(ninixw-preferences PUBLIC cxx_std_20)
target_include_directories(ninixw-preferences PUBLIC ${INCLUDE_DIRS})
target_link_libraries(ninixw-preferences PRIVATE ${FLTK_LIBRARIES} jsoncpp_lib fontlist)

add_executable(ninixw-msgbox ninixw-msgbox.cc msgbox.cxx)
set_target_properties(ninixw-msgbox PROPERTIES SUFFIX ".exe")
target_compile_options(ninixw-msgbox PUBLIC -O2 -Wall)
target_compile_features(ninixw-msgbox PUBLIC cxx_std_20)
target_include_directories(ninixw-msgbox PUBLIC ${INCLUDE_DIRS})
target_link_libraries(ninixw-msgbox PRIVATE ${FLTK_LIBRARIES})

add_executable(ninixw-scriptlog ninixw-scriptlog.cc)
set_target_properties(ninixw-scriptlog PROPERTIES SUFFIX ".exe")
target_compile_options(ninixw-scriptlog PUBLIC -O2 -Wall)
target_compile_features(ninixw-scriptlog PUBLIC cxx_std_20)
target_include_directories(ninixw-scriptlog PUBLIC ${INCLUDE_DIRS})
target_link_libraries(ninixw-scriptlog PRIVATE ${FLTK_LIBRARIES})
