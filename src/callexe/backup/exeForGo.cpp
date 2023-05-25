
#include <iostream>
#include <fstream>
#include <cassert>
#include <string>
#include "windows.h"

void writeTxtFile(std::string filePath, std::string txtInfo) {
    std::ofstream OutFile(filePath.data());
    OutFile << txtInfo;
    OutFile.close();
}
void readTxtFile(std::string filePath)
{
    std::ifstream infile;
    infile.open(filePath.data());   //将文件流对象与文件连接起来 
    assert(infile.is_open());   //若失败,则输出错误消息,并终止程序运行 

    std::string s;
    while (std::getline(infile, s))
    {
        std::cout <<"content: " << s << std::endl;
    }
    infile.close();
}

unsigned int taskID = 0;
unsigned int renderingTimes = 0;
void writeToJsonFile(std::string outputFilePath, unsigned int progress) {

    std::string statusJson = R"({
"rendering-ins":"jetty-scene-renderer",
"rendering-task":
    {
        "uuid":"rtrt88970-8990",
        "taskID":)" + std::to_string(taskID) + R"(,
        "name":"high-image-rendering",
        "phase":"finish",
        "progress":)" + std::to_string(progress) + R"(,
        "times":)" + std::to_string(renderingTimes) + R"(
    },
"rendering-status":"task:running"
})";

    writeTxtFile(outputFilePath, statusJson);
}
int main(int argc, char** argv)
{
    std::cout << "C++ exec Hello World!\n";
    std::cout << "argc:"<< argc <<"\n";
    for (auto i = 0; i < argc; ++i) {
        std::cout << "argv["<<i<<"]:" << argv[i] << "\n";
    }
    // params: scene res path, taskID, times
    //ins.exe .\static\sceneres\scene01 1002 9

    if (argc > 2) {
        taskID = std::stoi(argv[2]);
    }
    if (argc > 3) {
        renderingTimes = std::stoi(argv[3]);
    }
    std::cout << "taskID: " << taskID << "\n";
    std::cout << "renderingTimes: " << renderingTimes << "\n";

    std::string respath = ".\\renderingInfo.json";
    respath = std::string(argv[1]) + std::string("renderingInfo.json");
    std::cout << "\nrespath: " << respath << "\n";

    readTxtFile(respath);

    std::string outputFilePath = ".\\renderingStatus.json";
    outputFilePath = std::string(argv[1]) + std::string("renderingStatus.json");
    unsigned progress = 0;
    writeToJsonFile(outputFilePath, progress);

    for (auto i = 3; i > 0; --i) {
        std::cout << "countdown: " << i << " s.\n";
        progress += 25;
        writeToJsonFile(outputFilePath, progress);
        Sleep(2000);
    }
    writeToJsonFile(outputFilePath, 100);
    /*
    unsigned progress = 0;
    std::string statusJson = R"({
"rendering-ins":"jetty-scene-renderer",
"rendering-task":
    {
        "uuid":"rtrt88970-8990",
        "taskID":)" + std::to_string(taskID) + R"(,
        "name":"high-image-rendering",
        "phase":"finish",
        "progress":)" + std::to_string(progress) + R"(,
        "times":)" + std::to_string(renderingTimes) + R"(
    },
"rendering-status":"task:running"
})";
    //system("pause");
    for (auto i = 3; i > 0; --i) {
        std::cout << "countdown: " << i << " s.\n";
        Sleep(1000);
    }
    std::string outputFilePath = ".\\renderingStatus.json";
    outputFilePath = std::string(argv[1]) + std::string("renderingStatus.json");
    writeTxtFile(outputFilePath, statusJson);
    //*/
}