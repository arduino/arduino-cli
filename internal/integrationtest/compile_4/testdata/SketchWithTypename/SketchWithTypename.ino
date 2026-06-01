template< typename T >
  struct Foo{
    typedef T Bar;
};

void setup() {
  func();
}

void loop() {}

typename Foo<char>::Bar func(){

}
