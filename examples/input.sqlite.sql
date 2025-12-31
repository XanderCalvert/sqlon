CREATE TABLE "settings_color_duotone" (
    "colors" TEXT,
    "name" TEXT,
    "slug" TEXT
);
INSERT INTO "settings_color_duotone" ("colors", "name", "slug") VALUES ('["#02285b","#ffffff"]', 'Primary and White', 'primary-and-white');

CREATE TABLE "settings_color_gradients" (
    "gradient" TEXT,
    "name" TEXT,
    "slug" TEXT
);
INSERT INTO "settings_color_gradients" ("gradient", "name", "slug") VALUES ('linear-gradient(0deg, #e1236c 0%, #02285b 100%)', 'Primary to Secondary', 'primary-secondary');

CREATE TABLE "settings_color_palette" (
    "color" TEXT,
    "name" TEXT,
    "slug" TEXT
);
INSERT INTO "settings_color_palette" ("color", "name", "slug") VALUES ('#02285b', 'Primary', 'primary');
INSERT INTO "settings_color_palette" ("color", "name", "slug") VALUES ('#e1236c', 'Secondary', 'secondary');
INSERT INTO "settings_color_palette" ("color", "name", "slug") VALUES ('#ffa500', 'Tertiary', 'tertiary');
INSERT INTO "settings_color_palette" ("color", "name", "slug") VALUES ('#000', 'Black', 'black');
INSERT INTO "settings_color_palette" ("color", "name", "slug") VALUES ('#fff', 'White', 'white');
INSERT INTO "settings_color_palette" ("color", "name", "slug") VALUES ('#eee', 'grey', 'grey');

CREATE TABLE "settings_shadow_presets" (
    "name" TEXT,
    "shadow" TEXT,
    "slug" TEXT
);
INSERT INTO "settings_shadow_presets" ("name", "shadow", "slug") VALUES ('Faint', '0 2px 4px rgb(10, 10, 10, 0.1)', 'faint');
INSERT INTO "settings_shadow_presets" ("name", "shadow", "slug") VALUES ('Light', '0 0 10px rgb(10, 10, 10, 0.1)', 'light');
INSERT INTO "settings_shadow_presets" ("name", "shadow", "slug") VALUES ('Solid', '6px 6px 0 currentColor', 'solid');

CREATE TABLE "settings_spacing_spacingSizes" (
    "name" TEXT,
    "size" TEXT,
    "slug" TEXT
);
INSERT INTO "settings_spacing_spacingSizes" ("name", "size", "slug") VALUES ('Extra small', '0.5rem', 'filter-xs');
INSERT INTO "settings_spacing_spacingSizes" ("name", "size", "slug") VALUES ('Small', '1rem', 'filter-sm');
INSERT INTO "settings_spacing_spacingSizes" ("name", "size", "slug") VALUES ('Medium', 'clamp(1rem, 0.825rem + 0.675vw, 1.5rem)', 'filter-md');
INSERT INTO "settings_spacing_spacingSizes" ("name", "size", "slug") VALUES ('Large', 'clamp(1.5rem, 1.325rem + 0.675vw, 2rem)', 'filter-lg');
INSERT INTO "settings_spacing_spacingSizes" ("name", "size", "slug") VALUES ('X-Large', 'clamp(1.5rem, 1.151rem + 1.349vw, 2.5rem)', 'filter-xl');
INSERT INTO "settings_spacing_spacingSizes" ("name", "size", "slug") VALUES ('XXL', 'clamp(2rem, 1.476rem + 2.024vw, 3.5rem)', 'filter-xxl');
INSERT INTO "settings_spacing_spacingSizes" ("name", "size", "slug") VALUES ('Huge', 'clamp(2rem, 0.953rem + 4.047vw, 5rem)', 'filter-huge');

CREATE TABLE "settings_spacing_units" (
    "value" TEXT
);
INSERT INTO "settings_spacing_units" ("value") VALUES ('%');
INSERT INTO "settings_spacing_units" ("value") VALUES ('px');
INSERT INTO "settings_spacing_units" ("value") VALUES ('em');
INSERT INTO "settings_spacing_units" ("value") VALUES ('rem');
INSERT INTO "settings_spacing_units" ("value") VALUES ('vh');
INSERT INTO "settings_spacing_units" ("value") VALUES ('vw');

CREATE TABLE "settings_typography_fontFamilies" (
    "fontFace" TEXT,
    "fontFamily" TEXT,
    "name" TEXT,
    "slug" TEXT
);
INSERT INTO "settings_typography_fontFamilies" ("fontFace", "fontFamily", "name", "slug") VALUES ('[{"fontFamily":"Arial","fontStyle":"normal","fontWeight":"400","src":[]},{"fontFamily":"Arial","fontStyle":"italic","fontWeight":"400","src":[]}]', 'Arial', 'Arial', 'arial-font');
INSERT INTO "settings_typography_fontFamilies" ("fontFace", "fontFamily", "name", "slug") VALUES (NULL, 'Times', 'Times', 'times-font');

CREATE TABLE "settings_typography_fontSizes" (
    "name" TEXT,
    "size" TEXT,
    "slug" TEXT
);
INSERT INTO "settings_typography_fontSizes" ("name", "size", "slug") VALUES ('Extra small', '12px', 'filter-xs');
INSERT INTO "settings_typography_fontSizes" ("name", "size", "slug") VALUES ('Small', '16px', 'filter-sm');
INSERT INTO "settings_typography_fontSizes" ("name", "size", "slug") VALUES ('Base', '20px', 'filter-base');
INSERT INTO "settings_typography_fontSizes" ("name", "size", "slug") VALUES ('Medium', '24px', 'filter-md');
INSERT INTO "settings_typography_fontSizes" ("name", "size", "slug") VALUES ('Large', '32px', 'filter-lg');
INSERT INTO "settings_typography_fontSizes" ("name", "size", "slug") VALUES ('X-Large', '48px', 'filter-xl');
INSERT INTO "settings_typography_fontSizes" ("name", "size", "slug") VALUES ('XXL', '64px', 'filter-xxl');

CREATE TABLE "templateParts" (
    "area" TEXT,
    "name" TEXT,
    "title" TEXT
);
INSERT INTO "templateParts" ("area", "name", "title") VALUES ('footer', 'cpt-post', 'News Post - Footer');
