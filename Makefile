push:
	git push origin master

serve:
	mkdocs serve

publish:
	# if conflict, delete gh-pages branch first
	git push origin master
	mkdocs gh-deploy
