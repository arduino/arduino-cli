template< uint16_t nBuffSize >
  class Foo{
    public: 
  
    template< uint16_t N >
      Foo &operator +=( const Foo<N> &ref ){
        //...
        return *this;
    }
};

Foo<64> a;
Foo<32> b;

void setup(){
  a += b;
}

void loop(){}
