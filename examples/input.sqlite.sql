CREATE TABLE "settings_color_gradients" (
    "gradient" TEXT,
    "name" TEXT,
    "slug" TEXT
);
INSERT INTO "settings_color_gradients" ("gradient", "name", "slug") VALUES ('linear-gradient(0deg, #e1236c 0%, #02285b 100%)', 'Primary to Secondary', 'primary-secondary');

CREATE TABLE "settings_color_palette" (
    "slug" TEXT,
    "name" TEXT,
    "color" TEXT
);
INSERT INTO "settings_color_palette" ("slug", "name", "color") VALUES ('primary', 'Primary', '#02285b');
INSERT INTO "settings_color_palette" ("slug", "name", "color") VALUES ('secondary', 'Secondary', '#e1236c');
INSERT INTO "settings_color_palette" ("slug", "name", "color") VALUES ('tertiary', 'Tertiary', '#ffa500');
INSERT INTO "settings_color_palette" ("slug", "name", "color") VALUES ('black', 'Black', '#000');
INSERT INTO "settings_color_palette" ("slug", "name", "color") VALUES ('white', 'White', '#fff');
INSERT INTO "settings_color_palette" ("slug", "name", "color") VALUES ('grey', 'grey', '#eee');

CREATE TABLE "settings_color_duotone" (
    "name" TEXT,
    "slug" TEXT,
    "colors" TEXT
);
INSERT INTO "settings_color_duotone" ("name", "slug", "colors") VALUES ('Primary and White', 'primary-and-white', '["#02285b","#ffffff"]');

CREATE TABLE "settings_spacing_units" (
    "value" TEXT
);
INSERT INTO "settings_spacing_units" ("value") VALUES ('%');
INSERT INTO "settings_spacing_units" ("value") VALUES ('px');
INSERT INTO "settings_spacing_units" ("value") VALUES ('em');
INSERT INTO "settings_spacing_units" ("value") VALUES ('rem');
INSERT INTO "settings_spacing_units" ("value") VALUES ('vh');
INSERT INTO "settings_spacing_units" ("value") VALUES ('vw');

CREATE TABLE "settings_spacing_spacingSizes" (
    "slug" TEXT,
    "name" TEXT,
    "size" TEXT
);
INSERT INTO "settings_spacing_spacingSizes" ("slug", "name", "size") VALUES ('filter-xs', 'Extra small', '0.5rem');
INSERT INTO "settings_spacing_spacingSizes" ("slug", "name", "size") VALUES ('filter-sm', 'Small', '1rem');
INSERT INTO "settings_spacing_spacingSizes" ("slug", "name", "size") VALUES ('filter-md', 'Medium', 'clamp(1rem, 0.825rem + 0.675vw, 1.5rem)');
INSERT INTO "settings_spacing_spacingSizes" ("slug", "name", "size") VALUES ('filter-lg', 'Large', 'clamp(1.5rem, 1.325rem + 0.675vw, 2rem)');
INSERT INTO "settings_spacing_spacingSizes" ("slug", "name", "size") VALUES ('filter-xl', 'X-Large', 'clamp(1.5rem, 1.151rem + 1.349vw, 2.5rem)');
INSERT INTO "settings_spacing_spacingSizes" ("slug", "name", "size") VALUES ('filter-xxl', 'XXL', 'clamp(2rem, 1.476rem + 2.024vw, 3.5rem)');
INSERT INTO "settings_spacing_spacingSizes" ("slug", "name", "size") VALUES ('filter-huge', 'Huge', 'clamp(2rem, 0.953rem + 4.047vw, 5rem)');

CREATE TABLE "templateParts" (
    "area" TEXT,
    "name" TEXT,
    "title" TEXT
);
INSERT INTO "templateParts" ("area", "name", "title") VALUES ('footer', 'cpt-post', 'News Post - Footer');

CREATE TABLE "settings_shadow_presets" (
    "name" TEXT,
    "shadow" TEXT,
    "slug" TEXT
);
INSERT INTO "settings_shadow_presets" ("name", "shadow", "slug") VALUES ('Faint', '0 2px 4px rgb(10, 10, 10, 0.1)', 'faint');
INSERT INTO "settings_shadow_presets" ("name", "shadow", "slug") VALUES ('Light', '0 0 10px rgb(10, 10, 10, 0.1)', 'light');
INSERT INTO "settings_shadow_presets" ("name", "shadow", "slug") VALUES ('Solid', '6px 6px 0 currentColor', 'solid');

CREATE TABLE "settings_typography_fontFamilies" (
    "fontFace" TEXT,
    "fontFamily" TEXT,
    "name" TEXT,
    "slug" TEXT
);
INSERT INTO "settings_typography_fontFamilies" ("fontFace", "fontFamily", "name", "slug") VALUES ('[{"fontFamily":"Arial","fontStyle":"normal","fontWeight":"400","src":[]},{"fontFamily":"Arial","fontStyle":"italic","fontWeight":"400","src":[]}]', 'Arial', 'Arial', 'arial-font');
INSERT INTO "settings_typography_fontFamilies" ("fontFace", "fontFamily", "name", "slug") VALUES (NULL, 'Times', 'Times', 'times-font');

CREATE TABLE "settings_typography_fontSizes" (
    "size" TEXT,
    "slug" TEXT,
    "name" TEXT
);
INSERT INTO "settings_typography_fontSizes" ("size", "slug", "name") VALUES ('12px', 'filter-xs', 'Extra small');
INSERT INTO "settings_typography_fontSizes" ("size", "slug", "name") VALUES ('16px', 'filter-sm', 'Small');
INSERT INTO "settings_typography_fontSizes" ("size", "slug", "name") VALUES ('20px', 'filter-base', 'Base');
INSERT INTO "settings_typography_fontSizes" ("size", "slug", "name") VALUES ('24px', 'filter-md', 'Medium');
INSERT INTO "settings_typography_fontSizes" ("size", "slug", "name") VALUES ('32px', 'filter-lg', 'Large');
INSERT INTO "settings_typography_fontSizes" ("size", "slug", "name") VALUES ('48px', 'filter-xl', 'X-Large');
INSERT INTO "settings_typography_fontSizes" ("size", "slug", "name") VALUES ('64px', 'filter-xxl', 'XXL');
