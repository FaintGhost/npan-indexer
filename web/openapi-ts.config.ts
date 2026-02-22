import { defineConfig } from "@hey-api/openapi-ts";

export default defineConfig({
  input: "../api/openapi.yaml",
  output: "src/api/generated",
  plugins: [
    "@hey-api/typescript",
    {
      name: "zod",
      compatibilityVersion: 3,
      definitions: true,
      responses: true,
    },
  ],
});
