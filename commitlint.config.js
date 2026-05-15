const OPTIONAL_SCOPE_TYPES = ["ci", "refactor"];
const ALLOWED_SCOPES = [
  "booking-ui",
  "admin-ui",
  "server",
  "deps",
  "deps-dev",
  "i18n",
  "main",
];

module.exports = {
  extends: ["@commitlint/config-conventional"],
  plugins: [
    {
      rules: {
        "conditional-scope": (parsed) => {
          const { type, scope } = parsed;

          // allow release-please PR title
          if (type === "chore" && scope === "main") return [true];

          if (OPTIONAL_SCOPE_TYPES.includes(type)) {
            if (!scope) return [true];
            return [
              ALLOWED_SCOPES.includes(scope),
              `scope must be one of: ${ALLOWED_SCOPES.join(", ")}`,
            ];
          }
          if (!scope) return [false, "scope must not be empty"];
          return [
            ALLOWED_SCOPES.includes(scope),
            `scope must be one of: ${ALLOWED_SCOPES.join(", ")}`,
          ];
        },
      },
    },
  ],
  rules: {
    "header-max-length": [2, "always", 150],
    "scope-enum": [0],
    "scope-empty": [0],
    "conditional-scope": [2, "always"],
  },
};
