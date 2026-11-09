#!/usr/bin/env node

const Module = require("module");

const originalLoad = Module._load;

Module._load = function patchedLoad(request, parent, isMain) {
  const name = String(request);
  if (name.endsWith("../cli/next-test.js") || name.endsWith("/cli/next-test.js")) {
    return {
      SUPPORTED_TEST_RUNNERS_LIST: ["playwright"],
      nextTest: async () => {},
    };
  }
  return originalLoad.apply(this, arguments);
};

const nextBin = require.resolve("next/dist/bin/next", { paths: [process.cwd()] });

process.argv = [
  process.argv[0],
  nextBin,
  "dev",
  "-p",
  process.env.DEV_WEB_PORT || "9821",
  "--hostname",
  process.env.DEV_WEB_HOST || "127.0.0.1",
];

require(nextBin);
