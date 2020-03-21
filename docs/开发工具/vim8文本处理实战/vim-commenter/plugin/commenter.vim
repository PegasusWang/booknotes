" nnoremap gc :call g:commenter#ToggleComment()<cr>


" v:count1 表示获取命令的数字，并且默认1。:help v:count
" <		Note: The <C-U> is required to remove the line range that you
" get when typing ':' after a count.
nnoremap gc :<c-u>call g:commenter#ToggleComment(v:count1)<cr>


" 或者通过可视模式的多行选择实现多行注释
"
" vnoremap gc :<c-u>call g:commenter#ToggleComment( \ line("'>") - line("'<")+1)<cr>
