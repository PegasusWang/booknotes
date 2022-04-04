" if comment_str exists
function! g:commenter#HasCommentStr()
    if exists('g:commenter#comment_str')
        return 1
    endif
    echom "vim-commenter doesn't work for filetype " . &ft . " yet"
    return 0
endfunction

" 检测一个行区间中的最小缩进数
function! g:commenter#DetectMinIndent(start, end)
    let l:min_indent = -1
    let l:i = a:start 
    while l:i <= a:end
        if l:min_indent == -1 || indent(l:i) < l:min_indent
            let l:min_indent = indent(l:i)
        endif
        let l:i += 1
    endwhile
    return l:min_indent
endfunction


function! g:commenter#InsertOrRemoveComment(lnum, line, indent, is_insert)
    " 处理无缩进的情况
    let l:prefix = a:indent > 0 ? a:line[:a:indent-1] : ''
    if a:is_insert
        call setline(a:lnum, l:prefix . g:commenter#comment_str . a:line[a:indent:])
    else
        call setline(
          \ a:lnum, l:prefix . a:line[a:indent+len(g:commenter#comment_str):])
    endif
endfunction

function! g:commenter#ToggleComment(count)
    if !g:commenter#HasCommentStr()
        return
    endif
    let l:start = line('.')
    let l:end = l:start + a:count - 1
    if l:end > line('$') " stop at enf of file
        let l:end = line('$')
    endif

    let l:indent = g:commenter#DetectMinIndent(l:start, l:end)
    let l:lines = l:start == l:end ?
      \ [getline(l:start)] : getline(l:start, l:end) let l:cur_row = getcurpos()[1]
    let l:cur_col = getcurpos()[2]
    let l:lnum = l:start
    if l:lines[0][l:indent:l:indent+len(g:commenter#comment_str)-1] ==#
      \ g:commenter#comment_str
        let l:is_insert = 0
        let l:cur_offset = -len(g:commenter#comment_str)
    else
        let l:is_insert = 1
        let l:cur_offset = len(g:commenter#comment_str)
    endif
    for l:line in l:lines
        call g:commenter#InsertOrRemoveComment(l:lnum, l:line, l:indent, l:is_insert)
        let l:lnum += 1
    endfor
    call cursor(l:cur_row, l:cur_col+l:cur_offset)
endfunction

" Toggle comment
" function! g:commenter#ToggleComment()
"     if !g:commenter#HasCommentStr()
"         return
"     endif
"     let l:i = indent('.')
"     echo l:i
"     let l:line = getline('.')
"     let l:cur_row = getcurpos()[1]
"     let l:cur_col = getcurpos()[2]
"     " 处理无缩进情况
"     let l:prefix = l:i > 0 ? l:line[:l:i-1] : ''
"     if l:line[l:i:l:i+len(g:commenter#comment_str)-1] == g:commenter#comment_str
"         call setline('.', l:prefix . l:line[l:i+len(g:commenter#comment_str):])
"         let l:cur_offset = -len(g:commenter#comment_str)
"     else
"         call setline('.', l:prefix . g:commenter#comment_str . l:line[l:i:])
"         let l:cur_offset = len(g:commenter#comment_str)
"     endif
"     call cursor(l:cur_row, l:cur_col+l:cur_offset)
" endfunction
"
