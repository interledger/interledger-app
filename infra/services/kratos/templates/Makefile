format:
	@echo "Build styles"
	npx tailwindcss -m -i ./app.css -o ./minified.css -c ./tailwind.config.js
	@echo "Copy styles into templates"
	for file in ./recovery/valid/email.body ./recovery/invalid/email.body ./verification/valid/email.body ./verification/invalid/email.body; do \
		cat $$file.html > $$file.gotmpl;\
		echo "<style>" >> $$file.gotmpl;\
		cat ./minified.css >> $$file.gotmpl;\
		echo "\n</style>" >> $$file.gotmpl;\
	done
	
	@echo "Templates ready"