
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
