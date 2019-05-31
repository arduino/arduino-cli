void setup() { 
	myctagstestfunc(1,2,3,4); 
	test();
	test3();
	test5(1,2,3);
	test7();
	test8();
	test9(42, 42);
	test10(0,0,0);
}

void myctagstestfunc(int a,
int b,
int c,
int d) { }

void loop() {}

void 
test() {}

void 
// comment
test3() {}

void
test5(int a,
   int b,
   int c)
{

}

void /* comment */
test7() {}

void 
/* 
multi 
line 
comment 
*/
test8() {}

 void 
 /* comment */
 test9(int a, 
    int b) {} 

void test10(int a, // this 
			int b, // doesn't 
			int c  // work
			) {} 