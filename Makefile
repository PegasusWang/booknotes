push:
	git push origin master

serve:
	mkdocs serve

publish:
	# if conflict, delete gh-pages branch and site dir
	git push origin master
	mkdocs gh-deploy

clean:
	rm -rf site
