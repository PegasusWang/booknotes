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

