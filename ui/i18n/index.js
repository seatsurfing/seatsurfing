const i18n = {
  translations: {
    "en-GB": require("./translations.en-GB.json"),
    "en-US": require("./translations.en-US.json"),
    de: require("./translations.de.json"),
    et: require("./translations.et.json"),
    fi: require("./translations.fi.json"),
    fr: require("./translations.fr.json"),
    he: require("./translations.he.json"),
    hu: require("./translations.hu.json"),
    it: require("./translations.it.json"),
    nl: require("./translations.nl.json"),
    pl: require("./translations.pl.json"),
    pt: require("./translations.pt.json"),
    ro: require("./translations.ro.json"),
    es: require("./translations.es.json"),
    "zh-TW": require("./translations.zh-TW.json"),
  },
  defaultLang: "en-GB",
  useBrowserDefault: true,
  languageDataStore: "localStorage",
};

module.exports = i18n;
