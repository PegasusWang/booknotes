let animal_names = {
  \ 'cat': 'cat',
  \ 'dog': 'Dogson',
  \ 'parrot': 'Polly'
  \}

function animal_names.GetGreeting(animal)
  return self[a:animal] . ' says hello'
endfunction

echo animal_names.GetGreeting('cat')

let Hi = {animal -> animal . 'says hello'}
echo Hi('cat')

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


let animal = 'cat'
execute 'echo ' . animal

if has('python3')
  echom 'support python3'
endif

" expand 用于操作文件路径信息
echom 'current file' . expand('%:e')

if filereadable(expand('%'))
  echom 'Current file (' . expand('%:t') . ') is readable!'
endif

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
