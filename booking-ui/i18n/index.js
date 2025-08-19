var en = require("./translations.en.json");
var de = require("./translations.de.json");
var et = require("./translations.et.json");
var fr = require("./translations.fr.json");
var he = require("./translations.he.json");
var hu = require("./translations.hu.json");
var it = require("./translations.it.json");
var nl = require("./translations.nl.json");
var pt = require("./translations.pt.json");
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
    pt,
    ro,
  },
  defaultLang: "en",
  useBrowserDefault: true,
  languageDataStore: "localStorage",
};

module.exports = i18n;