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
    "scope-enum": ({ type, scope }) => {
      if (OPTIONAL_SCOPE_TYPES.includes(type) && !scope) return [true];
      return [
        ALLOWED_SCOPES.includes(scope),
        `scope must be one of: ${ALLOWED_SCOPES.join(", ")}`,
      ];
    },
    "scope-empty": ({ type, scope }) => {
      if (OPTIONAL_SCOPE_TYPES.includes(type)) return [true];
      return [!!scope, "scope must not be empty"];
    },
  },
};
