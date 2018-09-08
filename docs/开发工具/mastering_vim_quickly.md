# 5 Vim Concepts

常用模式(Modes):


- Normal mdoe
- insert mode
- command mode
- visual mode


# 11. Registers

`"ayy`  拷贝行到 a 寄存器 ，然后使用 `"ap` 粘贴，如果是大写的寄存器名称，会把内容 append 进去而不是替换

# 12. Buffers

buffer 是加载文件的内存区，直到保存文件才会从内存写入硬盘。
可以用 `:ls` 现实所有的 buffer，然后使用 `:5b` 打开序号为 5 的 buffer。
使用 `ctrl-6` 可以在两个 buffer 之间切换(依然是同一个窗口)。（当然不清楚为啥是 6 这个数字）


# 13. Windows, Tabs and Sessions

A window  is a viewport on a buffer, a Tab page is a collections of windows.
可以使用 `-o` 参数为每个打开的文件显示一个 window，比如 `vim -O file1.txt file2.txt`

使用 `:sp` `:vs` 可以横分屏和竖分屏

使用 `ctrl-w h` 也就是 h,j,k,l 可以在不同窗口之间切换。如果觉得麻烦可以映射

`nnoremap <C-H> <C-W><C-H>` 这样就可以使用 `ctrl+h` 来移动到左边的窗口了。

使用 `ctrl-w J` 大写字母可以移动当前窗口到最下方。
使用 `:only` 可以关闭所有其他的分屏

Session: 用来保存 vim 会话。 `:mksession ~/mysession.vim`，可以 `vim -S ~/mysession.vim` 或者 进入vim之后 `:source ~/mysession.vim` 启动。

# 14. Macros

宏：用来记录一系列操作行为，并且能够重放。

- 使用 qa 开始记录内容到寄存器 a 中
- 再次按下 q 退出宏的录制
- 使用 @a 使用宏
- 使用 @@ 重复上次执行的宏

宏每次执行都会重绘，如果觉得慢可以使用 set lazyredraw

# 15. The power of Visual modes
