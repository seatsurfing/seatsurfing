const HttpBackend = require('i18next-http-backend/cjs')
const ChainedBackend = require('i18next-chained-backend').default
const LocalStorageBackend = require('i18next-localstorage-backend').default
const isBrowser = typeof window !== 'undefined'
const isDev = process.env.NODE_ENV === 'development'

module.exports = {
  debug: isDev,
  browserLanguageDetection: true,
  nonExplicitSupportedLngs: true,
  backend: {
    backendOptions: [{
      expirationTime: isDev ? 0 : 60 * 60 * 1000, // 1 hour
    }, {
      loadPath: '/ui/locales/{{lng}}/{{ns}}.json',
    }], 
    backends: isBrowser ? [LocalStorageBackend, HttpBackend] : [],
  },
  i18n: {
    defaultLocale: 'default',
    locales: ['default', 'en', 'de', 'fr', 'it', 'he', 'hu', 'nl', 'ro'],
  },
  localeDetection: false,
  trailingSlash: true,
  use: isBrowser ? [ChainedBackend] : [],
}
