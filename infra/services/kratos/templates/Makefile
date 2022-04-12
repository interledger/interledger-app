format:
	@echo "Build styles"
	npx tailwindcss -m -i ./app.css -o ./minified.css -c ./tailwind.config.js
	@echo "Copy styles into templates"
	for file in ./recovery/valid/email.body ./recovery/invalid/email.body ./verification/valid/email.body ./verification/invalid/email.body; do \
		echo "<html lang='en'><head><meta charSet='utf-8' /><meta name='viewport' content='width=device-width, initial-scale=1' /><link rel='preconnect' href='https://fonts.googleapis.com' /><link rel='stylesheet' href='https://fonts.googleapis.com/css2?family=Source+Code+Pro&family=Overpass+Mono&family=Inter:wght@400;500&family=Poppins:wght@400;500&display=swap' /></head><body>" > $$file.gotmpl;\
		cat $$file.html >> $$file.gotmpl;\
		echo "<style>" >> $$file.gotmpl;\
		cat ./minified.css >> $$file.gotmpl;\
		echo "\n</style></body></html>" >> $$file.gotmpl;\
	done
	
	@echo "Templates ready"