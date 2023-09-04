class Foo {
public:
int blooper(int x) { return x+1; }
};

Foo foo;

void setup() {
  foo.blooper(1);
}

void loop() {
  foo.blooper(2);
}
