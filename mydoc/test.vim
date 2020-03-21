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
let g:is_dog  = 0
echo g:animal_name . (g:is_cat ? ' is a cat' : ' is something else' )

let g:is_dog  = 0
if !g:is_cat && !g:is_dog
  echo g:animal_name . 'is something else'
endif


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
" call remove(animals, 0, 1)

let mammals = ['dog', 'cat']
let birds = ['raven', 'parrot']
let animals = mammals + birds " add two list
call extend(mammals, birds) "追加 mammals
call sort(animals) " 字典序
let i = index(animals, 'parrot') " 查找索引
if empty(animals)
  echo 'no animals'
endif

call count(animals, 'cat') " 统计个数

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
