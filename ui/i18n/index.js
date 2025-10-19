const en_gb = require("./translations.en-GB.json");
const en_us = require("./translations.en-US.json");
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
    "en-GB": en_gb,
    "en-US": en_us,
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
  defaultLang: "en-GB",
  useBrowserDefault: true,
  languageDataStore: "localStorage",
};

module.exports = i18n;
