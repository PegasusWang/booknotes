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
