module.exports = {
  extends: ["@commitlint/config-conventional"],
  rules: {
    "header-max-length": [2, "always", 150],
    "scope-enum": [
      2,
      "always",
      ["booking-ui", "admin-ui", "core", "deps", "deps-dev", "i18n"],
    ],
  },
};
