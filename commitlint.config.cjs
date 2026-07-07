module.exports = {
  rules: {
    "header-max-length": [2, "always", 72],
    "type-enum": [
      2,
      "always",
      ["feat", "fix", "docs", "style", "refactor", "perf", "test", "build", "ci", "chore", "revert"],
    ],
    "subject-full-stop": [2, "never", "."],
    "subject-case": [0],  // Allow any case for subjects (e.g., "CI", "Go", etc.)
  },
};