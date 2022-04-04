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
