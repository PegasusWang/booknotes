# let 和 const 命令
### let

let 声明的变量只在所在的代码块中有效。 let 不存在变量提升，，一定要先声明后使用（暂时性死区）。
let 实际上为 js 新增了块级作用域，外层作用域无法读取内层作用域的变量。块级作用域的出现使得立即执行表达式不再必要了。
es5 规定函数不能在块级作用域定义，但是 es6 允许在其中声明函数，但应该尽量避免，而是使用函数表达式。

```js
for (let i = 0; i < 10; i++) {
    // ...
}
console.log(i)  // ReferenceError: i is not defined
```

### const
const 声明一个只读变量，一旦声明不可变。作用域和 let 相同


# 变量的解构赋值

### 数组解构赋值
按照『模式匹配』从数组和对象中抽取值。

```
let [a, b, c] = [1,2,3];

let [head, ...tail] = [1,2,3,4]
head // 1
tail // [2,3,4]

let [x,y, ...z] = ['a']
x // "a"
y // undefined, 解构不成功变量的值就是 undefined
z // []

// 不完全解构也可以成功。事实上，只要具备 Iterator 接口，都可以用数组形式解构赋值
let [x,y]=[1,2,3]
x //1
y //2
```

```
function* fibs() {
  let a=0;
  let b=1;
  while (true) {
    yield a;
    [a,b] = [b, a+b];
  }
}
```

```
// es6 内部使用严格相等运算符 ===
let [x=1] = [undefined];
x //1
let [x=1] = [null]
x //null，以为 null 不是严格等于 undefined，所以可以被赋值
```

注意如果默认值是一个表达式，他是惰性求值的。

### 对象解构赋值
先找到同名属性，然后赋值给对应的变量

```
let {foo, bar} = {foo:"aa", bar:"bb"};
foo // "aa"
bar // "bb"
```

用途：

- 1.交换变量值

```
let x=1;
let y=2;
[x, y] = [y, x]
```

- 2.从函数返回多个值

```
function example() {
  return [1,2,3]
}
let [a,b,c] = example()
function example() {
  return {
    foo: 1,
    bar: 2
  }
}
let {foo,bar} = example()
```

- 3.函数参数定义

```
function f([x,y,z]) {...}
f([1,2,3])
function f({x,y,z}) {...}
f({x:1, y:2, z:3})
```

- 4.提取 json 数据

```
let jsonData = {
  id: 42,
  status: "OK",
  data: [867, 5309]
};

let { id, status, data: number } = jsonData;

console.log(id, status, number);
// 42, "OK", [867, 5309]
```

- 5.参数默认值

```
jQuery.ajax = function (url, {
  async = true,
  beforeSend = function () {},
  cache = true,
  complete = function () {},
  crossDomain = false,
  global = true,
  // ... more config
}) {
  // ... do stuff
};
```

- 6.遍历 Map

```
var map = new Map();
map.set('first', 'hello');
map.set('second', 'world');

for (let [key, value] of map) {
  console.log(key + " is " + value);
}
// first is hello
// second is world”
```

- 7.输入模块的指定方法

```
const { SourceMapConsumer, SourceNode } = require("source-map");”
```

# 默认参数

### rest 参数(...变量名)

用于获取函数的多余参数

```
function add(...values) {
  let sum = 0;
  for (var val of values) {
    sum += val
  }
  return sum
}
add(2, 5, 3)  // 10

const sortNumbers = (...numbers) => numbers.sort();
```

### 箭头函数

es6 允许使用 (=>) 定义函数

```
var f = v => v;
// equal to
var f = function(v) {
  return v
}
var sum = (num1, num2) => num1 + num2;
var sum = (num1, num2) {
  return num1 + num2;
}


// 如果箭头函数代码块多于一条语句，需要用大括号括起来，并使用 return 返回
var sum = (num1, num2) => {return num1 + num2;}

//由于大括号被解释为代码块，所以如果返回一个对象，必须加上括号
let getTemplItem = id => ({id:id, name: "Temp"})

const isEven = n => n%2 == 0;
const square = n => n*n;

//一个用途是简化回调
[1,2,3].map(function(x) {
  return x * x;
})
[1,2,3].map(x => x * x);

//
var result = values.sort(function(a,b) {
  return a-b;
})

va result = values.sort((a,b) => a-b)
```


# Symbol

es6 引入了一种新的原始数据类型 Symbol，表示独一无二的值。

# Set 和 Map 结构

set : add, delete, has, clear; keys() values() entries() forEach()

```
const s=  new Set();
[1,2,3,4].forEach(x => s.add(x));
for (let i of s) {
  console.log(i);
}
```

WeakSet: 成员只能是对象；其中的对象都是弱引用。适合临时存放一组对象，以及存放跟对象绑定的信息。

Map: js 对象本质上是键值对的集合(Hash 结构)，但是传统上只能用字符串当做键，使用受限。

- size, set, get, has, delete, clear
- keys(), values(), entries(), forEach()

WeakMap: 只接受对象作为键名（null除外），WeakMap 键名指向的对象不计入垃圾回收机制。


# Proxy

用于修改某些操作的默认行为，等同于在语言层面修改。属于一种元编程。

# Promise 对象

一个容器，保存着某个未来才会结束的事件（通常是一个异步操作的结果）

# Iterator 和 forof 循环

Symbol.iterator 属性。具备原生 iterator 接口的数据结构如下：

Array, Map, Set, String, TypedArray, 函数的arguments 对象，NodeList 对象

```
let arr = ['a', 'b', 'c'];
let iter = arr[Symbol.iterator]();

iter.next() // { value: 'a', done: false }
iter.next() // { value: 'b', done: false }
iter.next() // { value: 'c', done: false }
iter.next() // { value: undefined, done: true }”
```

### for...of 循环
一个数据结构只要部署了 Symbol.iterator 属性，就被视为有 iterator接口，就能用
for...of 遍历它的成员。

```
var arr = ['a', 'b', 'c', 'd'];

for (let a in arr) {
  console.log(a); // 0 1 2 3
}

for (let a of arr) {
  console.log(a); // a b c d
}
```
for...in 循环读取键名，for...of 循环读取键值。

Set, Map 原生具有 iterator 接口，可以直接用 for...of

```
var engines = new Set(["Gecko", "Trident", "Webkit", "Webkit"]);
for (var e of engines) {
  console.log(e);
}
// Gecko
// Trident
// Webkit

var es6 = new Map();
es6.set("edition", 6);
es6.set("committee", "TC39");
es6.set("standard", "ECMA-262");
for (var [name, value] of es6) {
  console.log(name + ": " + value);
}
// edition: 6
// committee: TC39
// standard: ECMA-262”

```

# Generator 函数语法

```
function* helloWorldGenerator() {
  yield 'hello';
  yield 'world';
  return 'ending';
}

var hw = helloWorldGenerator(); //返回遍历器对象
```
不得不说和 python generator 很像。

async 函数返回一个 promise 对象。

# Class

之前 js 通过构造函数编写类，es6 提供了 class 语法糖。

```
function Point(x, y) {
  this.x = x;
  this.y = y;
}

Point.prototype.toString = function () {
  return '(' + this.x + ', ' + this.y + ')';
};

var p = new Point(1, 2);
```

```
class Point {
  constructor(x, y) {
    this.x = x;
    this.y = y;
  }

  toString() {   // 调用实例方法
    return '(' + this.x + ', ' + this.y + ')';
  }
}

class Point {
  // ...
}

typeof Point // "function" ，类的数据类型就是函数，类本身就指向构造函数，可以用 new 创建
Point === Point.prototype.constructor // true”
```

类和模块的内部默认是严格模式，无需手动指定。es6 实际上把整个语言升级到了严格模式。


### class 的继承

class 可以通过 extends 实现继承。子类必须在 constructor 中调用 super 方法。任何子类都有 constructor 方法（默认添加）。
调用 super 之后才能使用 this 关键字，否则会报错。

```
class Point {
}

class ColorPoint extends Point {
}
```

### Mixin 模式的实现
mixin 模式指的是，将多个类的接口『混入』『mix in』另一个类。

```
function mix(...mixins) {
  class Mix {}

  for (let mixin of mixins) {
    copyProperties(Mix, mixin);
    copyProperties(Mix.prototype, mixin.prototype);
  }

  return Mix;
}

function copyProperties(target, source) {
  for (let key of Reflect.ownKeys(source)) {
    if ( key !== "constructor"
      && key !== "prototype"
      && key !== "name"
    ) {
      let desc = Object.getOwnPropertyDescriptor(source, key);
      Object.defineProperty(target, key, desc);
    }
  }
}
```

# 修饰器 Decorator

```
@testable
class MyTestClass {
  //...
}

function testable(target) {
  taget.isTestable = true;
}
MyTestClass.isTestable // true


@decorator
class A {}

// 等同于
class A {}
A  = decorator(A) || A
```

修饰器只能用于类和类的方法， 无法用于函数，因为存在函数提升。类是不会提升的。
可以用高阶函数的形式修饰。

```
function doSomething(name) {
  console.log("Hello" + name);
}

function loggingDecorator(wrapped) {
  return function() {
    console.log("starting");
    const result = wrapped.apply(this, arguments);
    console.log("Finished");
    return result;
  };
}
const wrapped = loggingDecorator(doSomething);
```

core-decorator.js 一个第三方模块，提供了几个常见的修饰器。 @autobind， @readonly @override @deprecate, @suppressWarnings

### 修饰器实现 Mixin 模式

```
// mixins.js
export function mixins(...list) {
  return function(target) {
    Object.assign(target.prototype, ...list);
  };
}

import { mixins } from "./mixins";
const Foo = {
  foo() {
    console.log("foo");
  }
};
@mixins(Foo)
class MyClass {}

let obj = new MyClass();
obj.foo();
```

### Trait
也是一种修饰器，与Mixin类似，提供更多功能，比如防止同名方法冲突，排除混入某些方法，为混入方法起别名等。
traits-decorator 第三方模块

```
import { traits } from "traits-decorator";

class TFoo {
  foo() {
    console.log("foo");
  }
}

const TBar = {
  bar() {
    console.log("bar");
  }
};

@traits(TFoo, TBar)
class MyClass {}

let obj = new MyClass();
obj.foo();
obj.bar();
```

### Babel 转码器的支持

Babel 已经支持 Decorator。babel-core, babel-plugin-transform-decorators


# Module 的语法
es6 实现了模块，取代CommonJS 和 AMD 规范(这些是运行时确定)，成为浏览器和服务器通用方案。
es6 模块不是对象，而是通过export 命令显示指定输出的代码，再通过 import 导入。（编译时加载，静态加载）
es6 模块自动采用严格模式。

- 变量必须声明后再使用
- 函数的参数不能有同名属性，否则报错
- 不能使用 with 语句
- 不能对只读属性赋值，否则报错
- 不能使用前缀0表示八进制数，否则报错
- 不能删除不可删除的属性，否则报错
- 不能删除变量 delete prop ，会报错，只能删除属性 delete global[prop] 
-  eval 不会在它的外层作用域引入变量
-  eval 和 arguments 不能被重新赋值
-  arguments 不会自动反映函数参数的变化
- 不能使用 arguments.callee 
- 不能使用 arguments.caller 
- 禁止 this 指向全局对象
- 不能使用 fn.caller 和 fn.arguments 获取函数调用的堆栈
- 增加了保留字（比如 protected 、 static 和 interface ）

### export
export 规定模块的对外接口，import用于输入其他模块提供的功能。
一个模块是一个独立的文件，内部所有变量外部无法获取。可以使用 as 作为别名。
export 语句需要放在模块顶层。

```
// export 规定的对外接口，必须和模块内部变量建立一一对应关系
// 三种写法都可以

// 1
export var m = 1;
// 2
var m = 1;
export {m}
// 3
var n = 1;
export {n as m}
```

### import 

import  .. from '' 后边指定需要导入的位置，可以是相对或者绝对路径，js 后缀可以省略掉。
import 编译器期间执行的，静态执行所以无法使用表达式和变量。这些是只有运行时才能获得的语法结构。并且是单例的。

```
//整体加载。
import * as circle from './circle';

import { area, circumference } from './circle';

console.log('圆面积：' + area(4));
console.log('圆周长：' + circumference(14));

// wrong , 不允许运行时改变
circle.foo = 'hello'
```

### export default

给模块指定默认输出。export default 本质上就是输出一个叫做 default 的变量或者方法，然后系统运行你给它取任意名字。

### 模块可以继承

### 跨常量

```
export const A = 1;
export const B = 1;
```

### import() 函数

实现动态加载


# Module 的加载实现
