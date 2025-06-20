var en = require("./translations.en.json");
var de = require("./translations.de.json");
var et = require("./translations.et.json");
var fr = require("./translations.fr.json");
var he = require("./translations.he.json");
var hu = require("./translations.hu.json");
var it = require("./translations.it.json");
var nl = require("./translations.nl.json");
var ro = require("./translations.ro.json");

const i18n = {
  translations: {
    en,
    de,
    et,
    fr,
    he,
    hu,
    it,
    nl,
    ro,
  },
  defaultLang: "en",
  useBrowserDefault: true,
  // optional property will default to "query" if not set
  languageDataStore: "localStorage",
};

module.exports = i18n;