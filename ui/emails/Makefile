MJML=npx mjml
MJML_DIR=mjml
MJML_FILES=$(wildcard $(MJML_DIR)/*.mjml)
HTML_DIR=../../internal/mail/html

build: $(MJML_FILES)
	$(foreach file, $(MJML_FILES), $(MJML) $(file) --config.minify --config.minifyOptions='{"minifyCSS": true}' -o $(HTML_DIR)/$(basename $(notdir $(file))).html;)

.PHONY=build
