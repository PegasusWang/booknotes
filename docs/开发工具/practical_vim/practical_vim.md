> Do one thing, and do it well. - A principle of Unix philosopy

学习和使用vim有几年了，我的编辑器学习之路也该告一段落了，目前使用vim和一些插件基本能满足日常代码和文档(博客，笔记，代码文档)的编辑需求了，也无力折腾其他编辑器了。现在大部分时间都是在终端和vim里工作，我觉得在vim上花的时间还是值得的。本博客所有的文章都是在vim里写的，配合markdown插件可以在浏览器里即时预览。本博客是《Practical Vim》读书笔记，不是一本入门书，以一个个tip的形式组织，凝聚了作者的使用智慧，结合自己的使用经验做个简单的记录。

## CHAPTER 1: The Vim Way

- Tip1: Meet the Dot Command: 使用.重复上一个操作，可以视作宏。

- Tip2: Don't Repeat Yourself: 我们可以用A替代$a移动到行尾进行编辑。类似还有C,I,S,O等
- Tip3: Take One Step Back, Then Three Forward: 适当组合键位可以充分发挥.重复的优势。我们可以用f{char}来查找字母，";"可以用来重复上一次查找，之后用"."重复上次一操作，","可以用来返回上一次查找
- Tip4: Act, Repeat, Reverse: 善用undo操作。
- Tip5: Find and Replace by Hand: 使用`*`和`#`查找下一个或者上一个匹配。
- Tip6: Meet the Dot formula: One Keystroke to Move, One Keystroke to Execute.


## CHAPTER2: Normal Mode

- Tip7: Pause with Your Brush Off the Page: vim和其他编辑器一个显著不同就是默认不是编辑模式(insert mode)而是normal模式。实际上在编程中，大部分时间花在浏览，阅读，思考，组织和调整代码，而normal模式下众多命令能帮助我们大大提高效率。

- Tip8: Chunk your Undos: u能帮助我们撤销操作，合理在insert模式和normal模式切换能让undo操作更高效
- Tip9: Compose Repeatable Changes: vim is optimized for repetition。在vim里完成一个操作通常有多种方式，但是容易使用"."来重复的是比较好的方式。
- Tip10: Use Counts to Do Simple Arithmetic: vim很多命令都可以在前边加上数字表示重复几次命令，比如`<C-a>`对数字加1，如果用`10<C-a>`就是加上10。
- Tip11: Don't Count If You Can Repeat:  Keep your undo history clean.
- Tip12: Combine and Conquer:  Operator + Motion = Action，vim的强大在于组合命令的威力。比如d{motion}，"daw"是"delete a word"，比如"gU{motion}"的意思是Make {motion} text uppercase, "gUaw" 就是把一个单词转成大写，"gUap"就是把一个段落大写


## CHAPTER3: Insert Mode

- Tip13: Make Corrections Instantly from Insert Mode: 快速修正错误，`<C-h>`删除一个字符，`<C-w>`删除一个单词,`<C-u>`删除一行。实际上这些快捷键在和终端里是一样的，刻意替代退格键。

- Tip14: Get Back to Normal Mode: `<ESC>` 和 `<C-[>`都可以从insert模式切到normal模式，在insert模式下`<C-o>`是短暂切到normal模式之后切回insert模式，方便我们执行单个normal模式下的命令。配合上边快捷键我们就能在主键盘区域高效完成编辑操作。(把caps lock键改成ctrl)

- Tip15: Paste from a Register Without Leabing Insert Mode: `<C-r>{register}`，实现在insert模式下粘贴

- Tip16: Do Back-of-the-Envelope Calculations in Place:  vim的表达式寄存器(=)可以用来计算数值。`<C-r>=6*35<CR>`可以插入计算结果为210

- Tip17: Insert Unusual Characters By Character Code: 在插入模式下，可以用unicode码插入，比如 `<C-v>065` 将插入字符A

- Tip18: Insert Unusual Characters by Digraph: `:h digraph-table` ，也是用来插入特殊字符

- Tip19: Overwrite Existing Text with Replace Mode: `R`进入replace mode，平常我使用`r`进行单个字符修改，使用`R`可以从当前字符开始覆盖。


## CHAPTER4: Visual Mode

- Tip20: Grok Visual Mode:  Visual Mode可以实现字符、行、块选择，结合其他命令实现多种操作。

- Tip21: Define a Visual Selection: `v`字符选择, `V`行选择，`<C-v>`块选，`gv`重新执行最后一个visual mode的选择区域,`o`可以跳到选择的另一头，方便选择错误的时候两头调整。比如`vw`选中一个单词，`viw`无论你在单词的哪个位置都可以选中当前单词。`vi(`就更方便了，你可以在括号里的任何位置选中当前括号里所有内容，执行修改或者删除等操作，还有`vit`选中html标签里的内容，`vif`选中函数里的内容。

- Tip22: Repeat Line-Wise Visual Commands: 笔者经常使用visual mode选中需要缩进的代码块，然后执行`>`向右缩进，使用`.`重复缩进。

- Tip23: Prefer Operators to Visual Commands Where Possible: 优先使用操作符命令方便使用`.`重复命令

- Tip24: Edit Tabular Data with Visual-Block Mode:  vim的visual mode很适合编辑表格类文本。

- Tip25: Change Columns of Text: 在笔者使用nerd-commenter插件之前，实现多行注释都是用的多列编辑。

- Tip26: Append After a Ragged Visual Block: 结合$等，我们可以在vim里方便地实现给多行最后加上分号等功能。

## CHAPTER5: Command-Line Mode

- Tip27: Meet Vim's Command Line: 命令模式允许我们执行一个ex命令，搜索模式或者表达式

- Tip28: Execute a Command on One or More Consecutive Lines: ex模式下可以同时对多行进行操作，比如`:1,3p`打出前三行

- Tip29: Duplicate or Move Lines Using `:t` and `:m` Commands: `:t`是`:c`的简化命令，执行一样的功能。比如`:6copy.`把第六行拷贝到当前行下边，能直接用`:6t.`。`:m`用来执行移动行的功能，语法形式`:[range]move {address}`

- Tip30: Run Normal Mode Commands Across a Range: 再来看之前的一个例子，如果想给多行最后添加分号，我们刻意操作一行`A,<ESC>`表示在行尾插入一个逗号，之后我们在visual模式下选中多行，执行`:normal.`就能用`.`对这些行执行相同命令，或者执行`:normal A,`，也能给多行末尾添加逗号。结合多行选择和normal模式下的命令我们可以实现很多有用的操作。

- Tip31: Repeat the Last Ex Command: 我们经常使用`.`来重复normal模式的上一条命令，重复ex命令要使用`@:`

- Tip32: Tab-Complete Your Ex Commands: 和在shell里一样，可以使用tab补全ex命令，也可以使用`<C-d>`列出所有候选项，使用`<S-Tab>`回滚到上一个候选项。

- Tip33: Insert the Current Word at the Command Prompt: 在命令模式下，vim依然知道光标位置，使用`<C-r><C-w>`可以插入当前光标所在单词。我觉得这个命令在你想替换一个超长单词的时候比较有用，貌似对中文支持不行。

- Tip34: Recall Commands from History: 使用`:`进入命令模式以后就能使用上下键或者`<C-p><C-n>`上翻下翻命令，列出所有搜索历史命令执行`q/`，列出所有历史Ex命令执行`q:`，`C-f`从命令行模式切换到命令行窗口。


- Tip35: Run Commands in the Shell: 命令行模式前边加上`!`就可以执行shell命令，比如`:!ls`，注意内置的`:ls`显示的是buffer列表。也可以用`:shell`启动交互式shell会话，exit退出。`:read !{cmd}`可以使shell的命令写到buffer,`:write {cmd}`把buffer里的内容作为shell命令的标准输入。


## CHAPTER6: Manage Multiple Files

- Tip36: Track Open Files with the Buffer List: 在一次编辑会话中我们可以同时打开多个文件，vim使用buffer list管理他们。当我们在vim里编辑一个文件时，实际上是在编辑一个文件的内存表示，在vim术语里头叫做buffer。文件存储在硬盘，而buffers存在于内存中，大部分vim命令操作的是buffer，只有`:write`等少数命令会操作到硬盘中的文件。`:ls`列出所有buffer，`:bnext`或者`:bn`跳转到下一个buffer，`:bprev`或者`bp`跳到上一个，使用`C-^`来回切换，`:bd`移除当前buffer，同样还有`:bfirst`和`:blast`。如果觉得麻烦我可以可以映射命令`nnoremap <silent> [b :bprevious<CR>`

- Tip37: Group Buffers into a Collection with the ArgumentList: `:args`代表传入给vim命令的待编辑文件列表，我们也可以使用`:args file1 file2`命令组织待编辑的文件buffer。

- Tip38: Manage Hidden Files: 如果我们修改了buffer里的文件，必须保存才能切换到下一个文件，否则vim提示 "No write since last change (add ! to override)"。我们可以使用`:w`保存或者`:e!`放弃所有更改重新加载文件。

- Tip39: Divide Your Workspace into Split Windows: 分屏是笔者平常写代码的时候最最常用的功能，能让你同时浏览多个文件快速输理逻辑（所以平常写代码不要太长，限定80或者120列，否则分屏看起来比较痛苦）。可以用`<C-w>s`横分屏或者`<C-w>v`竖分屏，或者`:sp {file}`横分屏`:vsp {file}`竖分屏 。分频后，用`<C-w>h` `<C-w>j`等切换窗口，hjkl和移动命令对应，`<C-w>`循环切换。使用`<C-w>c`或者`:cl[ose]`关闭当前窗口，使用`<C-w>o`或`:on[ly]`只保留当前窗口。还有几个调整窗口大小的命令:`<C-w>=`重置所有窗口相等大小,`<C-w>_`当前窗口高度最大化，`<C-w>|`当前窗口宽度最大化。嗯，是不是命令太多头都大了，记住常用的几个天天使用就好，我现在就慢慢都习惯了，效率还是挺高的。

- Tip40: Organize Your Window Layouts with Tab Pages: buffer+分屏+tab组合起来实现了vim强大的多文件编辑功能。vim的tab和其他常用的编辑器有很大不同，大多数编辑器每打开一个文件就会生成一个tab，但是当我们用`:edit {file}`打开一个文件时，vim并不会自动创建新标签，而是使用新的buffer加载文件。这样的好处就是我们可以使用vim的tab来组织工作区，举个例子，笔者有时用一个vim开了三个tab（工作区），每个tab下我又分屏打开了多个文件，最多的时候能打开十来个文件。一个工作区编写后端代码，一个工作区查看前端代码，一个工作区用来修改测试脚本。当然这种情况比较少，打开这么多文件大脑负担会比较重，其实有了ctrlP等插件后定位和打开文件已经相当迅速了，在一个工作区中也能快速切换，而且开多了可能还会有性能问题。使用`:tabedit {filename}`打开新标签页，没有文件名打开一个包含空buffer的标签页。如果打开了多个窗口，使用`<C-w>T`会把当前窗口移到一个新的标签页，如果只使用一个窗口，使用`:close`关闭当前标签和窗口，或者`:tabclose`关闭当前标签页（不管有几个窗口），`:tabonly`只保留当前标签页面。使用`gt`和`gT`来回在标签页跳转，或者用`:tabn`和`:tabp`

## CHAPTER7: Open Files and Save Them to Disk

- Tip41: Open a File by Its Filepath Using `:edit` : 可以使用`:e {filepath}`在vim里打开文件。

- Tip42: Open a File by Its Filename Using `:finded {filename}` : 我更推荐ctrlP插件来查找文件，相当强大快捷。

- Tip43: Explore the File System with newtrw: 我更推荐nerdTree插件来实现文件树浏览。

- Tip44: Save Files to Nonexistent Directories: `:mkdir -p %:h`之后`:w`

- Tip45: Save a File as the Super User: 我们经常忘记sudo去编辑一个文件，结果保存的时候会提示readonly无法保存。`:w !sudo tee % > /dev/null`


## CHAPTER8: Navigate inside Fileds with Motions


- Tip46: Keep your Fingers on the Home Row: vim的很多设计让我们集中在主键盘区，从而避免的手指的来回移动。最明显的就是不使用上下左右，而是使用"hjkl"，习惯了以后效率会很高。

- Tip47: Distinguish Betwween Real Lines and Display Lines: 当一行过长时，vim会分几行显示，实际上这几行叫做显示行，看起来是多行实际上没有换行符。我们用`j`和`k`在真实行之间向下或者向上移动，而对于显示行，我们使用`gj`和`gk`来移动。实际上还有几个类似的命令都可以在前面加上g来操作显示行。比如`g0`和g`g&`移动到显示行的行首和行尾

- Tip48: Move Word-Wise: 文本编辑一个常用操作就是在单词之间移动，最常用的几个
    - w: 移到下一个单词首部, 助记(for-)Word
    - b: 移到上一个单词首部, 助记(Back-)word
    - e: 移到下一个单词尾部
    - ge: 移到上一个单词尾部

- Tip49: Find by Character: 一个常用操作是搜索一行里的字符`f{char}`，`;`重复搜索跳转到下一个相同字符，`,`跳转到上一个搜索字符位置。

    - f{char}: 移到char的下一次出现
    - F{char}: 移到char的上一次出现
    - t{char}: 移到(To)char出现的位置的前一个字符，我们可以用`dt{char}`删除直到字符char的下一次出现，比如我经常用来删除一个括号里的内容，`dt)`我经常用来删除右括号前的所有字符
    - T{char}: 移到上一个char的前一个字符位置


- Tip50: Search to Navigate: `/{word}<CR>`我们可以用命令模式中的`/`来执行搜索

- Tip51: Trace your Selection with Precision Text Objects: "Text objects"允许哦我们直接和括号，引用，xml标签等直接交互。比如我们想直接选定括号里的内容，使用`vi}`，"Text objects"通常用来操作结构化的数据，

- Tip52: Delete Around, or Change Inside: 使用text objects我们能实现在一个单词内部实现操作当前单词的需求，而不用移到单词首部

    - iw: current word
    - aw: current word plus one space, 比如我们可以用`daw`在一个单词的任意位置删除当前单词，而不用移到首部
    - is: current sentence
    - as: current sentence plus one space
    - ip: current paragraph
    - ap: current paragraph plus one blank line

- Tip53: Mark Your Place and snap Back to it :  在vim里可以给位置做标记从而快速跳转。`m{mark}`和 `mark  分别标记和跳转。有几个常用的跳转:

    - ``    当前文件的上一个跳转
    - `.    Location of Last change
    - `^    Location of last insertion

- Tip54: Jump Between Matching Parentheses: 最常用的就是使用%在括号之间跳转，我们还可以用vim插件matchit增强功能


## CHAPTER 9: Navigate Betwween Files With Jumps

- Tip55: Traverse the Jump List: `:jumps`能列出跳转历史，`<C-o>`和`<C-i>`能切换跳转

    - [count]G     跳转到行数
    - %     跳转到对应的括号
    - ( / )     跳转到上一个/下一个句子
    - { / }    跳转到上一个/下一个段落
    - H/M/L    H/M/L分别跳转到屏幕顶部/中间/尾部
    - gf      跳转到文件
    - <C-]>    跳转到当前关键字的第一次出现，经常用来查找定义
    - `{mark}`   跳转到标记

- Tip56: Traverse the Change List: vim会记录当前buffer所有改变，用`:changes`查看。一般用diff工具我觉得更好看点。

- Tip57: Jump to the Filename Under the Cursor:  不如直接用ctrlP插件快速打开文件，不用记忆命令了，vim命令够多了.

- Tip58: Snap Betwween Files Global Marks: 全局标记可以在文件中来回跳转，`m{char}`char为大写就是全局标记。


## CHAPTER 10: Copy and Paste

- Tip59: Delete, Yank, and Put with Vim's Unnamed Register: 我们可以用`xp`来反转字母，用`ddp`反转两行，这些都是使用的无名寄存器。

- Tip60: Grok Vim's Registers: vim提供了多个寄存器，我们可以用`"{register}`指定使用哪个寄存器，否则使用无名寄存器("")，所以使用`""p`等价直接用p命令。System Clipboard("+)和Selection("*) Registers

- Tip61: Replace a Visual Selection with a Register: 直接visual选中后使用粘贴命令p

- Tip62: Paste from a Register: 在insert模式下可以用`<C-r>"`

- Tip63: Interact with the System Clipboard: 粘贴的时候使用`:paste`能避免自动缩进，映射一个命令`:set pastetoggle=<F2>`比较方便，最好放到你的vimrc里。

## CHAPTER 11: Macros

- Tip64: Record and Execute a Macro: 宏经常用来对多行执行相同操作，我们可以录制宏并执行它。`q{register}`录制宏，之后q退出录制，用`:reg {register}`查看宏内容，用`@{register}`执行宏，使用`:[range]norm! @a`对多行执行宏。

- Tip65: Normalize, Strike, Abort: 黄金法则：录制宏的时候，保证每个命令都是可重复的。

- Tip66: Play Back with a Count; `.`重复命令的缺陷是无法指定次数，这个缺陷可以使用宏解决。

- Tip67: Repeat a Change on Contiguous Lines: 可以用块选择后执行`normal @a`针对多行执行宏。

- Tip68: Append Commands to a Macro: `q{char}`当char是大写的时候会把命令append到已经存在的对应的小写字母寄存器中。

- Tip69: Act Upon a Collection of Files: 需要结合argdo命令针对多个文件执行宏，笔者没用过。

- Tip70: Evaluate an Iterator to number Items in a List: 结合简单的vim脚本语句和宏能实现类似于给每行加上自增行号的需求。

- Tip71: Edit the Contents of a Macro: `put {register}`可以输出当前宏命令，编辑以后yank到寄存器里。


## CHAPTER 12: Matching Patterns and Literals

- Tip72: Tune the Case Sensitivity of Search Patterns: 全局设定忽略大小写`:set ignorecase`，对于单独的搜索在最后加上`\c`忽略大小写，
- Tip73: Use the \v Pattern Switch for Regex Searches: vim的正则语法接近POSIX而不是Perl的。使用`\v`后缀能切换vim的正则引擎更接近Perl,Python or Ruby.

- Tip74: Use the \V Literal Switch for Verbatim Searches: 使用`\V`意味着其后的pattern只有反斜线有特殊含义。如果你想匹配正则，使用 `\v`，如果你想精确匹配文本，使用`\V`

- Tip75: Use Parentheses to Capture Submatches: 使用括号匹配子组。`/\v<(\w+)\_s+\1>`可以匹配重复单词

- Tip76: Stake the Boundaries of a Word: `<`和`>`作为单词分界符，比如可以用`\<hello\>`精确匹配hello

- Tip77: Stake the Boundaries of a Match: 可以使用`\zs`和`\ze`截断匹配，只高亮需要的部分。

- Tip78: Escape Problem Characters: vim提供了escape函数来实现重复的转义操作

## CHAPTER 13: Search

- Tip79: Meet the Search Command: 命令模式输入`/`进入搜索模式并从当前光标位置往后搜索，`?`反向搜索。n和N分别是跳转到下一个和上一个搜索位置

- Tip80: Highlight Search Matches: 高亮搜索结果`:set hlsearch`，静默搜索`:se nohlsearch`

- Tip81: Preview the First Match Before Execution: `:set incsearch`增量搜索，每按下一个字符就会显示搜索到的单词。搜索中使用`<C-r><C-w>`补全。

- Tip82: Count the Matches for the Current Pattern:  `:% s/words//gn` 可以统计出有多少匹配到的words，实际上就是使用的替换模式加上标志n

- Tip83: Offset the Cursor to the End of a Search Match: 使用`/words/e<CR>`可以把光标放到搜索匹配的结尾

- Tip84: Operate on a Complete Search Match: vim没有自带比较好的实现，不过能通过插件textobj-lastpat实现。

- Tip85: Create Complex Patterns by Iterating upon Search History: 可以编辑search历史不断修正search的结果

- Tip86: Search for the Current Visual Selection: 在normal模式下可以用 * 找到下一个匹配，这个不支持块选的区域，可以通过脚本实现

## CHAPTER 14: Substitution

- Tip87: Meet the Substitute Command: 语法形式`:[range]s[ubstitute]/{pattern}/{string}/[flags]`，flags包含g(全局)，c(确认提示)，n(次数提示)，&(复用上一个flag)

- Tip88: Find and Replace Every Match in a File: 默认搜索只操作当前行第一个搜索，加flag g以后可以操作所有当前行的匹配，加上range之后我们可以操作所有当前文件的匹配。

- Tip89: Eyeball Each Substitution: 替换过程中使用flag c每次替换vim都会有提示，比较安全

- Tip90: Reuse the Last Search Pattern: Pressing <C-r>/ at the command line pastes the contents of the last search register in place.

- Tip91: Replace with the Contents of a Register: We can insert the contents of a register by typing <C-r>{register}。如果待搜索的内容过长，实际上可以在normal模式下复制到寄存器里，然后插入到替换命令里头。

- Tip92: Repeat the Previous Substitute Command:  We can always specify a new range and replay the substitution using the :&& command.

- Tip93: Rearrange CSV Fields Using Submatches: 可以用`\{num}`在替换语句里使用`/`搜索命令捕获的分组。

- Tip94: Perform Arithmetic on the Replacement: 这个是比较高级的用法，替换语句可以是一个vim脚本的表达式。

- Tip95: Swap Two or More Words : 我们可以创建一个vim脚本变量字典来执行替换操作。Abolish.vim插件提供了非常便捷的语法支持替换操作: `:%S/{man,dog}/{dog,man}/g`

- Tip96: Find and Replace Across Multiple Files: 需要结合argdo命令多文件操作，我觉得不如用sed命令方便点。

## CHAPTER 15: Global Commands

- Tip97: Meet the Global Command: The :global command allows us to run an Ex command on each line that matches a particular pattern. `:[range] global[!] /{pattern}/ [cmd]`

- Tip98: Delete Lines Containing a Pattern: `/\v\<\/?\w+>` `:g//d`可以删除文件中的html标签行，

- Tip99: Collect TODO Items in a Register: `:g/TODO`打出所有含有TODO的行，global模式命令执行echo。

- Tip100:  Alphabetize the Properties of Each Rule in a CSS File: `:g/{/ .+1,/}/-1 sort`


## CHAPTER 16: Index and Navigate Source Code with ctags

- Tip101: Meet ctags: ubuntu install ctags `sudo apt-get install exuberant-ctags`, mac `brew install ctags`

- Tip102: Configure Vim to Work with ctags:  生成tag命令`:!ctags -R`,映射一个`:nnoremap <f5> :!ctags -R<CR>`。

- Tip103: Navigate Keyword Definitions with Vim’s Tag Navigation Commands: 有了tags文件就可以用`<C-]>`跳转了。笔者目前写python居多，使用的是python-mode里集成的rope插件，用来跳转，查找定义等非常方便。
