import { createOpenRouter } from "@openrouter/ai-sdk-provider";
import { getEnvVar } from "./utils";

export const openrouter = createOpenRouter({
  apiKey: getEnvVar("OPENROUTER_API_KEY"),
});

export const model = openrouter("moonshotai/kimi-k2.6");

export const MAX_ITERATIONS = 15;
