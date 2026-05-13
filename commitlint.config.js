const OPTIONAL_SCOPE_TYPES = ["ci", "refactor"];
const ALLOWED_SCOPES = [
  "booking-ui",
  "admin-ui",
  "server",
  "deps",
  "deps-dev",
  "i18n",
];

module.exports = {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "header-max-length": [2, "always", 150],
    "scope-enum": async ({ type, scope }) => {
      if (OPTIONAL_SCOPE_TYPES.includes(type) && !scope) {
        return [0, "always", ALLOWED_SCOPES];
      }
      return [2, "always", ALLOWED_SCOPES];
    },
    "scope-empty": async ({ type }) => {
      if (OPTIONAL_SCOPE_TYPES.includes(type)) {
        return [0, "never"];
      }
      return [2, "never"];
    },
  },
};
