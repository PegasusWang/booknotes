# 3. 插件管理

分析 vim 性能： 使用 `-startuptime` 和 :profile 来分析性能

`:map g` 将显示所有 g 开头映射

`noremap <c-u> :w<cr>` 使用 ctrl-s 保存文件


```vim
let mapleader = "\<space>"
noremap <leader>w :w<cr>
noremap <leader>n :NERDTreeToggle<cr>
```

# 4. 理解文本

本章介绍了几个插件

- YouCompelteMe
- Exuberant Ctags
- Gundo 撤销树浏览

# 5. 构建、测试和运行

- tpope/vim-fugitive

使用 vimdiff 进行代码冲突合并。

```
<<<<<<<< [LOCAL commit/branch]
[LOCAL change]
||||||| merged commong ancestors
[BASE - closest common ancestor]
========
[REMOTE change]
>>>>>>>> [REMOTE commit/branch]
```

- vim-tmux-navigator
- vim-test 执行单测。 :TestNearest :TestFile :TestSuite :TestLast

# 6. 用正则表达式和宏重构代码

介绍了使用正则和 宏录制来进行重构

# 7. 定制自己的 vim

- airline 状态栏

vimrc 中手动折叠记号 `{{{` 。使用 zM 关闭所有折叠。

# 8. Vimscript

vimscript 使用一个双引号作为注释，有些命令后边不能用注释(和字符串定义冲突)

### 定义

```
" 使用 set 为内部选项赋值
set background=dark
" 非内部变量使用 let
let animal_name = 'Miss'
let is_cat = 1 " bool 使用1/0
"vim 中变量和作用域是通过前缀实现的
let g:animal_name = 'miss'
```

作用域规则：

- g为全局作用域（若未指定作用域，则默认为全局作用域）。
- v为Vim所定义的全局作用域。
- l为局部作用域（在函数内部，若未指定 作用域，则默认为这个作用域）。
- b表示当前缓冲区。
- w表示当前窗口。
- t表示当前标签页。
- s表示使用:source'd执行的Vim脚本文件中的局部文件作用域。
- a表示函数的参数。

```
" 寄存器设置同样使用 let
let @a = 'cats'
" vim选项(使用 set 修改的那些变量) 另一种访问方式是加上&
let &ignorecase = 0
" 整数可以进行 (+, - , *, /)，字符串拼接使用点 .
let g:cat = g:animal . ' is a cat'
```

### 打印

```
" 使用 echo 输出变量到状态栏
echo g:animal_name
" 使用 echomsg 可以记录输出结果，使用 :messages 就可以查看
echom 'here is another message'
```

### 条件表达式

```
let g:animal_kind = 'cat'
if g:animal_kind == 'cat'
  echo g:animal_name . ' is a cat'
elseif g:animal_kind == 'dog'
  echo g:animal_name . ' is a dog'
else
  echo g:animal_name . ' is something else'
endif

" vim 还支持三元运算符
let g:is_cat  = 0
echo g:animal_name . (g:is_cat ? ' is a cat' : ' is something else' )

" 逻辑运算符 与&& 或|| 非!
let g:is_dog  = 0
if !g:is_cat && !g:is_dog
  echo g:animal_name . 'is something else'
endif
```

字符串比较:

- == ，是否忽略大小写取决于用户设置
- ==? 忽略大小写比较
- ==# 考虑大小写
- =~ 使左操作数和右操作数(正则)匹配
- !~ 和 =~ 相反

### 列表

```
" vimscript 列表，操作类似 python
let animals = ['cat', 'dog', 'parrot']
let cat = animals[0]
let dog = animals[1]
let parrot = animals[-1] " 获取最后一个元素
let slice = animals[1:]
let slice2 = animals[0:1] "NOTE:注意 vim 索引包含右区间元素

call add(animals, 'octopus') " add element
let animals = add(animals, 'octopus')

call insert(animals, 'bobcat') " 插入元素，支持索引
call insert(animals, 'raven', 2)

unlet animals[2] " 删除索引 2 的元素
call remove(animals, -1) " 删除最后一个元素
let bobcat = remove(animals, 0)
unlet animals[1:]
call remove(animals, 0, 1)

let mammals = ['dog', 'cat']
let birds = ['raven', 'parrot']
let animals = mammals + birds " add two list
call extend(mammals, birds) "追加 mammals
call sort(animals) " 字典序
let i = index(animals, 'parrot') " 查找索引
if empty(animals)
  echo 'no animals'
endif

count(animals, 'cat') " 统计个数
```

### 字典

```
" :help dict
let animal_names = {
  \ 'cat': 'cat',
  \ 'dog': 'Dogson',
  \ 'parrot': 'Polly'
  \}

let cat_name = animal_names['cat']
let cat_name = animal_names.cat
let animal_names['raven'] = 'Raven'
let animal_names['raven2'] = 'Raven2'
unlet animal_names['raven2']
let raven = remove(animal_names, 'raven')
call extend(animal_names, {'bobcat': 'Sir'})

if !empty(animal_names)
  echo len(animal_names)
endif

if has_key(animal_names, 'cat')
  echo 'Cat is:' . animal_names['cat']
endif
```

### 循环

列表和字典都可以用 for 遍历

```
" iter list
for animal in animals
  echo animal
endfor

" iter dict by key
for animal in keys(animal_names)
  echo animal . animal_names[animal]
endfor

" iter dict by items
for [animal, name] in items(animal_names)
  echo animal . name
endfor

" break 和 continue
let animals = ['dot', 'cat', 'parrot']
for animal in animals
  if animal == 'cat'
    " 如果一个字符串内部使用单引号，需要用两个单引号
    echo 'It ''s a cat ! breaking!'
    break " continue
  endif
  echo 'Looking at a ' . animal . ', it'' s not a cat yet...'
endfor

" while
while !empty(animals)
  echo remove(animals, 0)
endwhile

" wihle 中同样可以使用 break 和 continue
while len(animals) > 0
  let animal = remove(animals, 0)
  if animal == 'dog'
    echo 'a dog, breaking'!
    break
  endif
endwhile
```

### 函数

```
" vim中自己定义的函数需要大写开头，除非函数在脚本作用域或者命名空间中
function Hello(animal)
  " 函数内部访问参数需要 a 作用域
  echo a:animal . ' hello!'
endfunction
call Hello("cat")

" 注意加上了!, 防止重复 source 报错重定义
function! Hello2(animal)
  return a:animal . ' says hello!'
endfunction
echo Hello2("dog")

" 可变参数
function! Hello3(...)
  echo a:1 . ' syas hi to ' . a:2
  " 获取所有参数
  echo a:000
endfunction
call Hello3('cat', 'dog')

" 结合固定参数和可变参数
function! Hello4(animal, ...)
  echo a:animal . 'says hi to ' . a:1
endfunction
call Hello4('cat', 'dog')

" 最好在函数中只使用局部作用域
function! s:AnimalGreeting(animal)
  echo a:animal . 'says hi!'
endfunction
" function! s:AnimalGreeting(animal)
"   echo a:animal . 'says hello!'
" endfunction
```

### 类

通过在字典上定义方法来实现

```
let animal_names = {
  \ 'cat': 'cat',
  \ 'dog': 'Dogson',
  \ 'parrot': 'Polly'
  \}

function animal_names.GetGreeting(animal)
  return self[a:animal] . ' says hello'
endfunction

echo animal_names.GetGreeting('cat')
```

### lambda 表达式
```
let Hi = {animal -> animal . 'says hello'}
echo Hi('cat')
```

### 高阶函数：映射(map)和过滤(filter)

map/filter 第一个参数是 list/dict，第二个是 function

```
let animal_names = {
  \ 'cat': 'cat',
  \ 'dog': 'Dogson',
  \ 'parrot': 'Polly'
  \}

func IsProperName(name)
  if a:name == 'cat'
    return 1
  endif
  return 0
endfunction

call filter(animal_names, 'IsProperName(v:val)')
echo animal_names

"filter 第二个参数也可以是函数引用，vim 支持引用一个函数o
let IsProperName2 = function('IsProperName')
echo IsProperName2('catt')
" 通过函数应用可以把函数作为参数传递给其他函数
function FunctionCaller(func, arg)
  return a:func(a:arg)
endfunction

echo FunctionCaller(IsProperName2, 'cat')

"filter 第二个参数也可以是函数引用，不过需要修改一下参数为两个，字典的key/val
function IsProperNameKeyValue(key, value)
  if a:value == 'cat'
    return 1
  endif
  return 0
endfunction

call filter(animal_names, function('IsProperNameKeyValue'))
echo animal_names

"filter 作用在列表时， v:key 表示索引，v:val 表示元素值

call map(animal_names,
  \ {key, val -> IsProperName(val) ? val : 'Miss ' .val})
```

### 与 vim 交互

execute 会把它的参数解析为一条vim命令执行，比如

```
let animal = 'cat'
execute 'echo ' . animal
```

normal 用于执行按键序列，与正常模式操作相似

```
" search cat and delete。
normal /cat<cr>dw
" normal 会遵守用户按键映射，忽略可以用 nomal!
nomal! /cat<cr>dw
```

silent 可以隐藏(比如execute)的输出。还可以隐藏外部命令输出，并禁止弹出对话框

```
silent echo animal . ' hello'
silent ececute 'echo ' . animal
silent !echo 'this is running in a shell'
```

has 检查 vim 是否支持某个功能

```
if has('python3')
  echom 'support python3'
endif
```

### 文件相关命令

```
" expand 用于操作文件路径信息
echom 'current file' . expand('%:e')

if filereadable(expand('%'))
  echom 'Current file (' . expand('%:t') . ') is readable!'
endif
```
- :p 展开为完整路径
- :h 路径头
- :t 路径结尾
- :r 根路径
- :e 文件扩展名

### 输入提示

两种方式：

- confirm 弹出一个对话框，读者从中选择多个答案
- input 支持自由形式的文本输入

```
let answer = confirm('is cat your favorite animal?', "&yes\n&no")
" yes 1, no 2，表示序号
echo answer

let animal = input("what is you favorite animal? ")
echo "\n"
echo 'Your favorite is a ' . animal

" 注意为了防止键盘映射中, 剩余字符被 input 获取，使用 inputsave/inputrestore
function AskAnimalName()
  call inputsave()
  let name = input('what is your favorite?')
  call inputrestore()
  return name
endfunction
nnoremap <leader>a = :let name= AskAnimalName()<cr>:echo name<cr>
```

编程风格参考：《google vimscript style guide》

### 编写一个插件

一般的 vim 插件目录结构

- autoload/ 保存插件延迟加载内容
- colors/目录用于保存配色。
- compiler/目录用于保存编译器相关的功能（针对不同语言）。
- doc/为文档目录。
- ftdetect/目录用于保存（针对不同文件类型的）文件类型检测设置。
- ftplugin/目录用于保存（针对不同文件类型的）文件类型相关的代码。
- indent/目录用于保存（针对不同文件类型的）缩进相关的设置。
- plugin/目录用于保存插件的核心功能。
- syntax/目录用于保存（针对不同语言的）语言语法分组。

用一个 vim-commenter 示例如何编写 vim plugin。
使用 helptags 可以生成帮助文档索引。
发布插件到 github 上之后就可以使用插件管理器安装了。

# 9. Neovim


