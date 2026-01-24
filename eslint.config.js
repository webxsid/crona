import tseslint from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import prettier from "eslint-config-prettier";

export default [
  {
    files: ["**/*.ts"],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        project: [
          "./tsconfig.json",
          "./packages/*/tsconfig.json",
          "./tests/*/tsconfig.json",
        ],
        sourceType: "module",
      },
    },
    plugins: {
      "@typescript-eslint": tseslint,
    },
    rules: {
      // Type safety
      "@typescript-eslint/no-unused-vars": [
        "warn",
        { argsIgnorePattern: "^_" },
      ],
      "@typescript-eslint/no-explicit-any": "error",

      // Code quality
      "no-console": "off",
      "no-throw-literal": "error",

      // Design discipline (important for kernel)
      "max-params": ["warn", 4],
      "complexity": ["warn", 20],

      // TS prefers explicit intent
      "@typescript-eslint/consistent-type-imports": "warn",
    },
  },

  // Disable rules that fight formatters
  prettier,
];
