void setup() {
   // put your setup code here, to run once:
 }
 
 void loop() {
   // put your main code here, to run repeatedly:
   test2();
   test4();
   test6();
   test7();
   test10();
 }

 extern "C" {
   void test2() {}
 }
  
 extern "C" 
 {
   void test4() {}
 }
  
 extern    "C"    
 
 {
   void test6() {}
 }

 // this function should not have C linkage
 void test7() {}

 extern    "C"     void test10() {
   
 };