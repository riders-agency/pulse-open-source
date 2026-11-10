import tsPlugin from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";

const browserGlobals = {
  Blob: "readonly",
  CustomEvent: "readonly",
  Event: "readonly",
  File: "readonly",
  FormData: "readonly",
  Headers: "readonly",
  HTMLElement: "readonly",
  HTMLInputElement: "readonly",
  HTMLTextAreaElement: "readonly",
  InputEvent: "readonly",
  KeyboardEvent: "readonly",
  MouseEvent: "readonly",
  Request: "readonly",
  Response: "readonly",
  URL: "readonly",
  URLSearchParams: "readonly",
  WebSocket: "readonly",
  clearInterval: "readonly",
  clearTimeout: "readonly",
  console: "readonly",
  crypto: "readonly",
  document: "readonly",
  fetch: "readonly",
  localStorage: "readonly",
  location: "readonly",
  navigator: "readonly",
  process: "readonly",
  sessionStorage: "readonly",
  setInterval: "readonly",
  setTimeout: "readonly",
  window: "readonly",
};

export default [
  {
    ignores: [
      ".next/**",
      "next-env.d.ts",
      "node_modules/**",
      "playwright-report/**",
      "test-results/**",
    ],
  },
  {
    files: ["**/*.{js,jsx,ts,tsx}"],
    languageOptions: {
      parser: tsParser,
      parserOptions: {
        ecmaFeatures: {
          jsx: true,
        },
        ecmaVersion: "latest",
        sourceType: "module",
      },
      globals: browserGlobals,
    },
    plugins: {
      "@typescript-eslint": tsPlugin,
    },
    rules: {
      "@typescript-eslint/no-unused-vars": [
        "warn",
        {
          argsIgnorePattern: "^_",
          caughtErrorsIgnorePattern: "^_",
          ignoreRestSiblings: true,
          varsIgnorePattern: "^_",
        },
      ],
      "no-undef": "off",
      "no-unused-vars": "off",
      "no-var": "error",
      "prefer-const": "warn",
    },
  },
];
