# REQUIRES is left out on purpose: "main" then depends on every component,
# which works both with the component manager (espressif__esp_lvgl_port...)
# and with local copies of the components.
idf_component_register(
    SRCS "main.c" "board.c" "audio.c" "net.c" "portal.c" "api.c" "store.c"
         "ui_theme.c" "ui.c" "ui_study.c"
    INCLUDE_DIRS "."
    EMBED_FILES "fonts/ui_regular.ttf" "fonts/ui_bold.ttf"
)
spiffs_create_partition_image(storage ../storage_image FLASH_IN_PROJECT)
