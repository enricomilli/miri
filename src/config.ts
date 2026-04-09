import { createOpenRouter } from "@openrouter/ai-sdk-provider";

export const openrouter = createOpenRouter({
	apiKey: process.env.OPENROUTER_API_KEY ?? "",
});

export const model = openrouter("minimax/minimax-m2.7");

export const MAX_ITERATIONS = 15;
