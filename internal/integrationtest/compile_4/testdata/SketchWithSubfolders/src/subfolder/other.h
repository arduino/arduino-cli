#ifndef other__h
#define other__h
#include <testlib3.h>

class MyClass {
  public:
    MyClass();
    void init ( Stream *controllerStream );

  private:
    Stream *controllerStream;
};
#endif
