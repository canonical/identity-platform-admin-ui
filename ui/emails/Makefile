MJML=npx mjml
MJML_DIR=mjml
MJML_FILES=$(wildcard $(MJML_DIR)/*.mjml)
IMAGES_DIR_NAME=images
MJML_IMAGES=$(MJML_DIR)/$(IMAGES_DIR_NAME)
HTML_DIR=../../internal/mail/html
IMAGES_DIR=$(HTML_DIR)/$(IMAGES_DIR_NAME)

images:
	rsync -a --delete $(MJML_IMAGES)/ $(IMAGES_DIR)

.PHONY=images

build: $(MJML_FILES) images
	$(foreach file, $(MJML_FILES), $(MJML) $(file) --config.minify --config.minifyOptions='{"minifyCSS": true}' -o $(HTML_DIR)/$(basename $(notdir $(file))).html;)

.PHONY=build
