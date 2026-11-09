cmake_minimum_required(VERSION 3.22)
project(riders_pulse_desktop LANGUAGES CXX)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_AUTOMOC ON)

find_package(Qt6 COMPONENTS Core Network Gui REQUIRED)

add_executable(riders-pulse-desktop
  src/main.cpp
)

target_link_libraries(riders-pulse-desktop PRIVATE
  Qt6::Core
  Qt6::Network
  Qt6::Gui
)
