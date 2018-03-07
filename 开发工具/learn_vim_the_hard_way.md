<http://learnvimscriptthehardway.onefloweroneworld.com/> 读书笔记，自定义vim 配置或者一些简单的 vim 插件需求。

# 设置选项

主要有两种选项：布尔选项（值为"on"或"off"）和键值选项。

-   布尔选项设置取消`:set number` `:set nonumber`
-   查看当前值: `:set number?`
-   键值选项: `:set numberwidth=10`

# 基本映射

`:map - x`让减号成为删除

# 模式映射

你可以使用nmap、vmap和imap命令分别指定映射仅在normal、visual、insert模式有效。

-   `:nmap - dd`
-   `:nunmap - dd`

insert 模式: 比如我们想实现 ctrl+d 在 insert 模式下删除一行

-   `:imap <c-d> dd` 你会发现不起作用，因为没有进入 normal 模式
-   `:imap <c-d> <esc>dd` 在 insert 模式下输入 ctrl+d 从 esc 进入 normal 模式然后删除一行，但是没有回到 insert 模式
-   `:imap <c-d> <esc>ddi` 完成

# 精确映射

`*map`系列命令的一个缺点就是存在递归的危险。另外一个是如果你安装一个插件，插件 映射了同一个按键为不同的行为，两者冲突，有一个映射就无效了。
每一个\_map系列的命令都有个对应的\`\_noremap\`命令，包括：noremap/nnoremap、 vnoremap和inoremap。这些命令将不递归解释映射的内容。
该何时使用这些非递归的映射命令呢？ 答案是： 任何时候 。

# Leaders

    :let mapleader = ","
    :nnoremap <leader>d dd

`:let maplocalleader = "\\"` Vim有另外一个“leader”成为“local leader“ã这个leader用于那些只对某类文件 （如Python文件、HTML文件）而设置的映射。

# 编辑你的Vimrc文件

    " 编辑映射
    :nnoremap <leader>ev :vsplit $MYVIMRC<cr>

    " 重读映射配置
    :nnoremap <leader>sv :source $MYVIMRC<cr>

# Abbreviations

Vim有个称为"abbreviations"的特性，与映射有点类似，但是它用于insert、replace和 command模式。

-   用来纠错：`:iabbrev adn and` 在输入adn之后输入空格键，Vim会将其替换为and。
-   用来缩写：`:iabbrev @@    steve@stevelosh.com`

# 更多的Mappings

看一个复杂的,一个有趣的mappings！你自己可以先试试。进入normal模式，移动光标至一个单词， 输入<leader>"。Vim将那个单词用双引号包围！
`:nnoremap <leader>" viw<esc>a"<esc>hbi"<esc>lel`

# 锻炼你的手指

让我们先创建一个mapping，这个mapping会为你的左手减轻很多负担。
`:inoremap jk <esc>`
重新学习一个mapping的窍门就是强制将之前的按键设置为不可用，强迫自己使用新的mapping。执行下面的命令：
`:inoremap <esc> <nop>`

# 本地缓冲区的选项设置和映射

nnoremap命令中的<buffer>告诉Vim这个映射只在定义它的那个缓冲区中有效：
`:nnoremap <buffer> <leader>x dd`

# 自动命令

`:autocmd BufNewFile * :write`
创建空文件就写入到磁盘

    :autocmd BufNewFile * :write
             ^          ^ ^
             |          | |
             |          | 要执行的命令
             |          |
             |          用于事件过滤的“模式（pattern）”
             |
             要监听的“事件”

`:autocmd BufWritePre *.html :normal gg=G` BufWritePre，这个事件会在你保存任何字符到文件之前触发。保存 html 自动缩进

多个事件：
`:autocmd BufWritePre,BufRead *.html :normal gg=G`

FileType事件:
这个事件会在Vim设置一个缓冲区的filetype的时候触发。
通过 `<localleader>`给不同文件类型注释

    :autocmd FileType javascript nnoremap <buffer> <localleader>c I//<esc>
    :autocmd FileType python     nnoremap <buffer> <localleader>c I#<esc>

# 本地缓冲区缩写

`:iabbrev <buffer> --- &mdash;`

自动命令和缩写:
使用本地缓冲区的缩写和自动命令来创建一个简单的“snippet”系统。

    :autocmd FileType python     :iabbrev <buffer> iff if:<left>
    :autocmd FileType javascript :iabbrev <buffer> iff if ()<left>

# 自动命令组

Vim可能创建两个完全一样的自动命令,对于这个问题，Vim有一个解决方案。这个解决方案的第一步是将相关的自动命令收集起来放到一个已命名的组（groups）中。

    augroup filetype_html
        autocmd!
        autocmd FileType html nnoremap <buffer> <localleader>f Vatzf
    augroup END

当进入filetype_html这个组的时候，我们会立即清除这个组，然后定义一个自动命令，然后退出这个组。当我们再次加载~/.vimrc文件的时候，清除组命令会阻止Vim添加一个一模一样的自动命令。

# Operator-Pending映射

一个Operator（操作）就是一个命令，你可以在这个命令的后面输入一个Movement（移动）命令，然后Vim开始对文本执行前面的操作命令，这个操作命令会从你当前所在的位置开始执行，一直到这个移动命令会把你带到的位置结束。
常用到的Operator有d，y和c。例如：

| 按键  | 操作  | 移动     |
| --- | --- | ------ |
| dw  | 删除  | 到下一个单词 |
| ci( | 修改  | 在括号内   |
| yt, | 复制  | 到逗号    |

`:onoremap p i(`  dp Vim会删除括号内的所有文字

当你想搞清楚怎么定义一个新的operator-pending movement的时候，你可以从下面几个步骤来思考：

-   在光标所在的位置开始。
-   进入可视模式(charwise)。
-   ... 把映射的按键放到这里 ...
-   所有你想包含在movement中的文字都会被选中。
    你所要做的工作就是在第三步中填上合适的按键。
    <http://learnvimscriptthehardway.onefloweroneworld.com/chapters/15.html>

# 更多Operator-Pending映射

这个映射有些复杂。现在把你的光标放到文本中的某个位置（不要放到标题上）,然后敲击cih。Vim会删除光标所在章节的标题，然后保持在插入模式，这可以称为"修改所在的标题(change inside heading)"。
<http://learnvimscriptthehardway.onefloweroneworld.com/chapters/16.html>
`:onoremap ih :<c-u>execute "normal! ?^==\\+$\r:nohlsearch\rkvg_"<cr>`

# 状态条

    :set statusline=%f         " 文件的路径
    :set statusline+=%=        " 切换到右边
    :set statusline+=%l        " 当前行
    :set statusline+=/         " 分隔符
    :set statusline+=%L        " 总行数

# 负责任的编码

注释、分组、不要用缩写、练习并分享你的 vimrc

# 变量

    :let foo = "bar"
    :echo foo

你也可以将寄存器当作变量来读取和设置。执行下面的命令：

    :let @a = "hello!"

# 变量作用域

在两个分隔的窗口中分别打开两个不同的文件，然后在其中一个窗口中执行下面的命令：

    :let b:hello = "world"
    :echo b:hello

当你在变量名中使用b:，这相当于告诉Vim变量hello是当前缓冲区的本地变量。
当某个变量由一个字符和冒号开头，那么这就表示它是一个作用域变量。

# 条件语句

    :echom "foo" | echom "bar"

Vim会把它当作两个独立的命令。如果你看不到两行输出，执行:messages查看消息日志。

    :echom "hello" + 10
    :echom "10hello" + 10
    :echom "hello10" + 10

Vim不会把非空字符串当作"truthy"
如有必要，Vim将强制转换变量(和字面量)的类型。在解析10 + "20foo"时，Vim将把 "20foo"转换成一个整数(20)然后加到10上去。
以一个数字开头的字符串会被强制转换成数字，否则会转换成0
在所有的强制转换完成后，当if的判断条件等于非零整数时，Vim会执行if语句体。

    :if 0
    :    echom "if"
    :elseif "nope!"
    :    echom "elseif"
    :else
    :    echom "finally!"
    :endif

# 比较

==的行为取决于用户的设置。是否 ignorecase。不要依赖用户配置。所以怎样才能适应这荒谬的现实？好在Vim有额外两种比较操作符来处理这个问题。
`==?`是"无论你怎么设都大小写不敏感"比较操作符
`==#`是"无论你怎么设都大小写敏感"比较操作符。
当你比较整数时，这点小不同不会有什么影响。 不过，我还是建议每一次都使用大小写敏感的比较(即使不一定需要这么做)，好过该用的时候忘记用了

# 函数

没有作用域限制的Vimscript函数必须以一个大写字母开头！

    :function Meow()
    :  echom "Meow!"
    :endfunction

    :function GetMeow()
    :  return "Meow String!"
    :endfunction

调用函数

    :call Meow()
    :call GetMeow()

当你使用call时，返回值会被丢弃， 所以这种方法仅在函数具有副作用时才有用。

    :function TextwidthIsTooWide()
    :  if &l:textwidth ># 80
    :    return 1
    :  endif
    :endfunction

一个Vimscript函数不返回一个值，它隐式返回0

# 函数参数

    :function DisplayName(name)
    :  echom "Hello!  My name is:"
    :  echom a:name
    :endfunction

在写需要参数的Vimscript函数的时候，你总需要给参数加上前缀a:，来告诉Vim去参数作用域查找。

    :function Varg(...)
    :  echom a:0  " 0表示参数个数
    :  echom a:1
    :  echo a:000  " a:000将被设置为一个包括所有传递过来的额外参数的列表(list)。
    :endfunction

    :call Varg("a", "b")

可变参数

    :function Varg2(foo, ...)
    :  echom a:foo
    :  echom a:0
    :  echom a:1
    :  echo a:000
    :endfunction

    :call Varg2("a", "b", "c")

可以用具名参数引用。不能对参数变量重新赋值

# 数字

Vimscript有两种数值类型：Number和Float。一个Number是32位带符号整数。一个Float是浮点数。
计算过程中也会转换类型

# 字符串

写Vimscript的时候，确信你清楚写下的每一个变量的类型。如果需要改变变量类型，你就得使用一个函数显式改变它， 即使那不是必要的。不要依赖Vim的强制转换，毕竟世上没有后悔药。

-   连接：`:echom "Hello, " . "world"` ，

Vim不会把非空字符串当作"truthy"

# 字符串函数

-   strlen(长度): `:echom strlen("foo")`
-   split(切割)：`:echo split("one two three")` `:echo split("one,two,three", ",")`
-   join(连接):`echo join(["foo", "bar"], "...")`
-   大小写转换: `:echom tolower("Foo")` `:echom toupper("Foo")`

# Execute命令

execute命令用来把一个字符串当作Vimscript命令执行
`:execute "echom 'Hello, world!'"`
 无需担心 execute 的危险性。

# Normal命令

`:normal G`
Vim将把你的光标移到当前文件的最后一行，就像是在normal模式里按下G。假如G 已经被映射了，我们可以用:
`:normal! G`
在写Vim脚本时，你应该总是使用normal!，永不使用normal。不要信任用户在~/.vimrc中的映射。

# 执行normal!

`:execute "normal! gg/foo\<cr>dd"`
这将移动到文件的开头，查找foo的首次出现的地方，并删掉那一行。

`:execute "normal! mqA;\<esc>`q"\`

这个命令做了什么？让我们掰开来讲：

-   :execute "normal! ..."：执行命令序列，一如它们是在normal模式下输入的，忽略所有映射， 并替换转义字符串。
-   mq：保存当前位置到标记"q"。
-   A：移动到当前行的末尾并在最后一个字符后进入insert模式。
-   ;：我们现在位于insert模式，所以仅仅是写入了一个";"。
-   \\<esc>：这是一个表示Esc键的转义字符串序列，把我们带离insert模式。
-   \`q：回到标记"q"所在的位置。

看上去有点绕，不过它真的很有用：它在当前行的末尾补上一个分号并保持光标不动。 在写Javascript，C或其他以分号作为语句分隔符的语言时，一旦忘记加上分号，这样的映射将助你一臂之力。

# 基本的正则表达式

高亮: 在开始之前，先花点时间讲讲搜索高亮，这样我们可以让匹配的内容更明显。hlsearch让Vim高亮文件中所有匹配项，incsearch则令Vim在你正打着搜索内容时就高亮下一个匹配项
`:set hlsearch incsearch`

搜索: `:execute "normal! gg/print\<cr>"` 这将移动到文件顶部并开始搜索print，带我们到第一处匹配。

`:execute "normal! gg" . '/\vfor .+ in .+:' . "\<cr>"` 查找 python 的 for in 语句
我们又一次把正则表达式放在单独的字面量字符串里，而这次我们用\\v来引导模式。 这将告诉Vim使用它的"very magic"正则解析模式，而该模式就跟其他语言的非常相似。

# 实例研究：Grep 运算符(Operator)，第一部分

简明扼要地说：:grep ...将用你给的参数来运行一个外部的grep程序，解析结果，填充quickfix列表， 这样你就能在Vim里面跳转到对应结果。

让我们简化我们的目标为"创造一个映射来搜索光标下的词"。这有用而且应该更简单，所以我们能更快得到可运行的成果。 目前我们将映射它到<leader>g。
`:nnoremap <leader>g :grep -R something .<cr>`

首先我们需要搜索光标下的词，而不是something。执行下面的命令：

`:nnoremap <leader>g :grep -R <cword> .<cr>`
现在试一下。<cword>是一个Vim的command-line模式的特殊变量， Vim会在执行命令之前把它替换为"光标下面的那个词"。

你可以使用<cWORD>来得到大写形式(WORD)。
我们将调用参数用引号括起来(防止危险的字符串)。执行这个命令：

`:nnoremap <leader>g :grep -R '<cWORD>' .<cr>`

搜索部分还有一个问题。在that's上尝试这个映射。它不会工作，因为词里的单引号与grep命令的单引号发生了冲突！
为了解决问题，我们可以使用Vim的shellescape函数。 阅读:help escape()和:help shellescape()来看它是怎样工作的(真的很简单)。
因为shellescape()要求Vim字符串，我们需要用execute动态创建命令。 首先执行下面命令来转换:grep映射到:execute "..."形式：

`:nnoremap <leader>g :execute "grep -R '<cWORD>' ."<cr>`

试一下并确信它可以工作。如果不行，找出拼写错误并改正。 然后执行下面的使用了shellescape的命令。

`:nnoremap <leader>g :execute "grep -R " . shellescape("<cWORD>") . " ."<cr>`

在一般的词比如foo上执行这个命令试试。它可以工作。再到一个带单引号的词，比如that's，上试试看。 它还是不行！为什么会这样？
问题在于Vim在拓展命令行中的特殊变量，比如<cWORD>，的之前，就已经执行了shellescape()。 所以Vim shell-escaped了字面量字符串"<cWORD>"(什么都不做，除了给它添上一对单引号)并连接到我们的grep命令上。
注意引号也是输出字符串的一部分。Vim把它作为shell命令参数保护了起来。
为解决这个问题，我们将使用expand()函数来强制拓展<cWORD>为对应字符串， 抢在它被传递给shellescape之前。

`:nnoremap <leader>g :exe "grep -R " . shellescape(expand("<cWORD>")) . " ."<cr>`

在完成映射之前，还要处理一些小问题。首先，我们说过我们不想自动跳到第一个结果， 所以要用grep!替换掉grep。执行下面的命令：

`:nnoremap <leader>g :execute "grep! -R " . shellescape(expand("<cWORD>")) . " ."<cr>`

再一次试试，发现什么都没发生。Vim用结果填充了quickfix窗口，我们却无法打开。 执行下面的命令：

`:nnoremap <leader>g :execute "grep! -R " . shellescape(expand("<cWORD>")) . " ."<cr>:copen<cr>`

# 实例研究：Grep运算符(Operator)，第二部分

    " http://learnvimscriptthehardway.onefloweroneworld.com/chapters/33.html

    " 然后执行g@来以运算符的方式调用这个函数
    nnoremap <leader>g :set operatorfunc=GrepOperator<cr>g@
    " 还想要在visual模式下用到它。
    " 我们传递过去的visualMode()参数还没有讲过呢。 这个函数是Vim的内置函数，它返回一个单字符的字符串来表示visual模式的类型： "v"代表字符宽度(characterwise)，"V"代表行宽度(linewise)，Ctrl-v代表块宽度(blockwise)。
    vnoremap <leader>g :<c-u>call GrepOperator(visualmode())<cr>

    function! GrepOperator(type)
    	if a:type ==# 'v'
    		normal! `<v`>y
    	elseif a:type ==# 'char'
    		normal! `[v`]y
    	else
    		return
    	endif

    	silent execute "grep! -R " . shellescape(@@) . " ."
    	copen
    endfunction

# 实例研究：Grep运算符(Operator)，第三部分

    nnoremap <leader>g :set operatorfunc=GrepOperator<cr>g@
    vnoremap <leader>g :<c-u>call GrepOperator(visualmode())<cr>

    function! GrepOperator(type)
        let saved_unnamed_register = @@

        if a:type ==# 'v'
            normal! `<v`>y
        elseif a:type ==# 'char'
            normal! `[v`]y
        else
            return
        endif

        silent execute "grep! -R " . shellescape(@@) . " ."
        copen

        let @@ = saved_unnamed_register
    endfunction

我们在函数的开头和结尾加入了两个let语句。 第一个用一个变量保存@@中的内容，第二个则重新加载保存的内容。
当写Vim插件时，你总是应该尽量在修改之前保存原来的设置和寄存器值，并在之后加载回去。 这样你就避免了让用户陷入恐慌的可能。

我们的脚本在全局命名空间中创建了函数GrepOperator。 这大概不算什么大问题，但当你写Vimscript的时候，事前以免万一远好过事后万分歉意。

    nnoremap <leader>g :set operatorfunc=<SID>GrepOperator<cr>g@
    vnoremap <leader>g :<c-u>call <SID>GrepOperator(visualmode())<cr>

    function! s:GrepOperator(type)
        let saved_unnamed_register = @@

        if a:type ==# 'v'
            normal! `<v`>y
        elseif a:type ==# 'char'
            normal! `[v`]y
        else
            return
        endif

        silent execute "grep! -R " . shellescape(@@) . " ."
        copen

        let @@ = saved_unnamed_register
    endfunction

脚本的前三行已经被改变了。首先，我们在函数名前增加前缀s:，这样它就会处于当前脚本的命名空间。
我们也修改了映射，在GrepOperator前面添上<SID>，所以Vim才能找到这个函数。 如果我们不这样做，Vim会尝试在全局命名空间查找该函数，这是不可能找到的。

# 列表

-   嵌套 `:echo ['foo', [3, 'bar']]`
-   索引：`:echo [0, [1, 2]][1]`
-   切割：`:echo ['a', 'b', 'c', 'd', 'e'][0:2]`
-   连接: `:echo ['a', 'b'] + ['c']`
-   函数：add, len, get, index, join, reverse

# 循环

for 和 while

    :let c = 0

    :for i in [1, 2, 3, 4]
    :  let c += i
    :endfor

    :echom c

    :let c = 1
    :let total = 0

    :while c <= 4
    :  let total += c
    :  let c += 1
    :endwhile

    :echom total

# 字典

我们讲到的最后一种Vimscript类型将是字典。 Vimscript字典类似于Python中的dict，Ruby中的hash，和Javascript中的object。
字典用花括号创建。值是异质的，但键会被强制转换成字符串。

-   定义：`:echo {'a': 1, 100: 'foo'}   " {'a':1,'100':'foo'}`
-   索引：`:echo {'a': 1, 100: 'foo',}['a']` or `:echo {'a': 1, 100: 'foo',}.a` ，支持括号和点查找
-   赋值和添加


    :let foo = {'a': 1}
    :let foo.a = 100
    :let foo.b = 200
    :echo foo

-   移除: 你不能移除字典中不存在的项


    :let test = remove(foo, 'a')
    :unlet foo.b
    :echo foo
    :echo test

-   字典函数: get(), has_key(), items(), keys(), values()

# 切换

`:nnoremap <leader>N :setlocal number!<cr>` 在normal模式中按下<leader>N看看。Vim将会在开启和关闭行号显示之间切换。

    nnoremap <leader>q :call QuickfixToggle()<cr>

    let g:quickfix_is_open = 0

    function! QuickfixToggle()
        if g:quickfix_is_open
            cclose
            let g:quickfix_is_open = 0
            execute g:quickfix_return_to_window . "wincmd w"
        else
            let g:quickfix_return_to_window = winnr()
            copen
            let g:quickfix_is_open = 1
        endif
    endfunction

# 函数式编程

如果你用过Python，Ruby或Javascript，甚或Lisp，Scheme，Clojure或Haskell，
你应该会觉得把函数作为变量类型，用不可变的状态作为数据结构是平常的事。

### 不可变的数据结构

不幸的是，Vim没有类似于Clojure内置的vector和map那样的不可变集合， 不过通过一些辅助函数，我们可以在一定程度上模拟出来。

    function! Sorted(l)
        let new_list = deepcopy(a:l)
        call sort(new_list)
        return new_list
    endfunction

### 作为变量的函数

Vimscript支持使用变量储存函数，但是相关的语法有点愚钝。执行下面的命令：

    :let Myfunc = function("Append")
    :echo Myfunc([1, 2], 3)

### 高阶函数

    function! Mapped(fn, l)
        let new_list = deepcopy(a:l)
        call map(new_list, string(a:fn) . '(v:val)')
        return new_list
    endfunction

# 路径

### 绝对路径

    :echom expand('%')
    :echom expand('%:p')
    :echom fnamemodify('foo.txt', ':p')

### 列出文件

    :echo globpath('.', '*') " Vim将输出当前目录下所有的文件和文件夹
    :echo split(globpath('.', '**'), '\n')  " 你可以用**递归地列出文件。执行这个命令：

# 创建一个完整的插件
