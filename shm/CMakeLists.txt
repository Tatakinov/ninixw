cmake_minimum_required(VERSION 3.30)
project(ninixw-shm CXX)

SET(SRCS main.cc)
SET(INCLUDE_DIRS ${PROJECT_SOURCE_DIR})

if(WIN32)
    SET(SRCS shm_windows.cc)
else()
    SET(SRCS ${SRCS} shm_others.cc)
    SET(LIBS -lpthread)
endif()

add_executable(ninixw-shm ${SRCS})

set_target_properties(ninixw-shm PROPERTIES SUFFIX ".exe")

target_compile_options(ninixw-shm PUBLIC -O2 -Wall)
target_compile_features(ninixw-shm PUBLIC cxx_std_20)

target_include_directories(ninixw-shm PUBLIC ${INCLUDE_DIRS})

target_link_libraries(ninixw-shm PRIVATE ${LIBS})
