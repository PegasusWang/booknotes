# 1. 入门


```java
class HelloWorld {
	public static void main(String [] args) {
		System.out.println("hello");
	}
}
```

classpath 环境变量:java虚拟机需要运行一个类，会在定义的路径中寻找所需要的class文件
自动设置为当前目录。


java sourcecode -> class字节码文件(javac compile) -> jvm 将字节码文件加载到内存(类加载器) -> 执行


### package

java 通过包组织类，声明用 package

	import package.class
	import package.*   // import all classs
	java.util java.net java.io java.awt java.sql


### JVM java virtual Machine

解释自己的指令集（字节码）到 cpu 的指令集或者 os 的系统调用，jvm 只关心 class file。
类文件的组成包括 JVM 指令集，符号表以及一些补助信息。

### JRE java runtime environment

jre 目录有 bin and lib, 可以认为 bin 里就是 jvm, lib 市 jvm 需要的一些类库

### JKD java development kit

java 开发工具包。 jkd 包含 jre, jre 包含 jvm


# 2. java 基础

类或者接口用大骆驼命名法(Student, StudentName)，方法和变量小骆驼(dogName).

常量全部大写。(STU_NAME)

	单行注释 //   多行 /* */   文档注释(javadoc)


原码：二进制定点表示法

反码：正数反码和源码相同，负数的反码是对齐源码逐位取反，但符号位除外

补码：正数补码与原码相同，负数补码是其反码末尾加1.


计算机操作的时候， 都是用数据对应的补码来计算的。

类型：

基本数据类型：数值(byte,short,int,long,float,double)；字符(char)；布尔(boolean)

引用数据类型：类(class); 接口(interface); 数组；枚举类型(enum);注解(Annotation)

方法重载：同一个类中允许多个同名方法，只要他们的参数个数或者参数类型不同即可。与返回值类型无关


数组： int[] arr = new int[3]; int[] arr = new int[]{1,2,3} or int[] arr = {1,2,3}

包装类：Byte,Short,Int,Long,Float,Double;  大数类：BigInteger, BigDecimal; 字符类：Character


# 3。 面向对象

封装：属性和行为封装，对外隐藏细节。隔离变化、便于使用、复用性和安全性。

继承：扩展类功能，复用性和开发效率。使用 extends继承。java 只允许单继承(继承单个类)
  - 子类可以直接访问父类中的非私有属性和行为
  - 子类无法继承父类中私有的内容


多态：重写父类方法表现不同行为。子类中重写的方法需要和父类被重写的方法具有相同的方法名、参数列表和返回值类型.

java 中方法永远传递值，但是如果是引用类型，传递的就是对象的引用。

super关键字：用于访问父类的成员变量、方法、和构造方法。


- 区分重载（同一个类中同名不同参）
- 重写or 覆盖(override)，子类覆盖父类方法
	- 无法覆盖父类中的私有方法
	- 无法覆盖父类中 static 方法
	- 覆盖时，子类方法权限一定要大于等于父类方法权限(更开放权限)

子类构造函数默认会执行父类的构造函数，第一行有一个默认的隐士语句：super().
super() 语句必须要定义在子类构造函数的第一行，父类的初始化动作要先完成。

final: 可以修饰类(不可被继承）、变量(常量，只能被赋值一次）、 方法（不可以被覆盖）。

静态绑定：方法被static/private/final 三个关键字其中之一修饰，执行的静态绑定。

多态：允许一个父类类型的变量引用一个子类类型的对象。父类或接口的引用指向或者接收自己的子类对象。
提高了扩展性，前期定义的代码可以使用后期的内容，但是不能使用后期子类*特有*的内容

instanceof: 用于判断对象的具体类型，只能用于引用数据类型判断。通常在向下转型前用于健壮性判断。.
if (a instanceof Cat)


- 成员变量：编译和运行时都参考等号左边。
- 成员函数：编译看左边， 运行看右边。 动态绑定
- 静态函数: 编译和运行都看左边
