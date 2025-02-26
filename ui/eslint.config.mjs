import tseslint from "typescript-eslint";
import react from "eslint-plugin-react";
import prettier from "eslint-plugin-prettier";
import globals from "globals";
import path from "node:path";
import { fileURLToPath } from "node:url";
import js from "@eslint/js";
import { FlatCompat } from "@eslint/eslintrc";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);
const compat = new FlatCompat({
  baseDirectory: __dirname,
  recommendedConfig: js.configs.recommended,
  allConfig: js.configs.all,
});

export default [
  {
    files: ["src/**/*.ts", "src/**/*.tsx"],
  },
  {
    ignores: ["public", "dist"],
  },
  ...compat.extends(
    "eslint:recommended",
    "plugin:react/recommended",
    "plugin:prettier/recommended",
    "plugin:@typescript-eslint/recommended",
    "plugin:@typescript-eslint/recommended-requiring-type-checking",
    "plugin:@typescript-eslint/strict",
  ),
  {
    plugins: {
      "@typescript-eslint": tseslint.plugin,
      react,
      prettier,
    },
    languageOptions: {
      globals: {
        ...globals.node,
        ...globals.browser,
        ...globals.jest,
      },
      parser: tseslint.parser,
      ecmaVersion: 5,
      sourceType: "module",
      parserOptions: {
        project: true,
        tsconfigRootDir: __dirname,
        ecmaFeatures: {
          jsx: true,
        },
      },
    },
    settings: {
      react: {
        version: "detect",
      },
    },
    rules: {
      "linebreak-style": ["error", "unix"],
      semi: ["error", "always"],
      "object-curly-spacing": ["error", "always"],
      "@typescript-eslint/no-unused-vars": [
        "error",
        {
          argsIgnorePattern: "^_",
        },
      ],
      "@typescript-eslint/prefer-promise-reject-errors": "off",
      "react/react-in-jsx-scope": "off",
    },
  },
];
