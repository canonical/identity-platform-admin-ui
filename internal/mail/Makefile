MJML=npx mjml
MJML_FILES=$(wildcard mjml/*.mjml)
HTML_DIR=html

build:
	$(foreach file, $(MJML_FILES), $(MJML) $(file) --config.minify --config.useMjmlConfigOptions -o $(HTML_DIR)/$(basename $(notdir $(file))).html;)

.PHONY=build
