cmake_minimum_required(VERSION 3.30)
project(ninixw-proxy CXX)

SET(SRCS main.cc)
SET(INCLUDE_DIRS ${PROJECT_SOURCE_DIR})

if(WIN32)
    SET(SRCS lib_windows.cc)
else()
    find_package(PkgConfig REQUIRED)
    pkg_check_modules(LIBFFI REQUIRED libffi)
    SET(SRCS ${SRCS} lib_others.cc lib_kawari8_others.cc lib_pasta_others.cc)
    SET(INCLUDE_DIRS ${INCLUDE_DIRS} ${LIBFFI_INCLUDE_DIRS})
    SET(LIBS ${LIBFFI_LIBRARIES} -ldl)
endif()

add_executable(ninixw-proxy ${SRCS})

set_target_properties(ninixw-proxy PROPERTIES SUFFIX ".exe")

target_compile_options(ninixw-proxy PUBLIC -O2 -Wall)
target_compile_features(ninixw-proxy PUBLIC cxx_std_20)

target_include_directories(ninixw-proxy PUBLIC ${INCLUDE_DIRS})

target_link_libraries(ninixw-proxy PRIVATE ${LIBS})
