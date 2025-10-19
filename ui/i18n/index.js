const en_gb = require("./translations.en-gb.json");
const en_us = require("./translations.en-us.json");
const de = require("./translations.de.json");
const et = require("./translations.et.json");
const fr = require("./translations.fr.json");
const he = require("./translations.he.json");
const hu = require("./translations.hu.json");
const it = require("./translations.it.json");
const nl = require("./translations.nl.json");
const pl = require("./translations.pl.json");
const pt = require("./translations.pt.json");
const ro = require("./translations.ro.json");
const es = require("./translations.es.json");

const i18n = {
  translations: {
    "en-gb": en_gb,
    "en-us": en_us,
    de,
    et,
    fr,
    he,
    hu,
    it,
    nl,
    pl,
    pt,
    ro,
    es,
  },
  defaultLang: "en-gb",
  useBrowserDefault: true,
  languageDataStore: "localStorage",
};

module.exports = i18n;
