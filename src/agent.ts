import { type ModelMessage, streamText } from "ai";
import { MAX_ITERATIONS, model } from "./config";
import { exec } from "./executor";
import { parseTerminalBlocks } from "./parser";
import { SYSTEM_PROMPT } from "./prompts";

export async function runAgent() {
	const messages: ModelMessage[] = [{ role: "system", content: SYSTEM_PROMPT }];

	while (true) {
		const userInput = prompt(`\x1b[36m╭─You\n╰─λ\x1b[0m`);
		if (userInput === null) break;
		if (!userInput.trim()) continue;

		messages.push({ role: "user", content: userInput });

		for (let i = 0; i < MAX_ITERATIONS; i++) {
			const userMessage = messages[messages.length - 1];

			const result = streamText({ model, messages });

			let fullText = "";

			process.stdout.write("\n\x1b[35m──Miri\n\x1b[0m ");
			for await (const chunk of result.textStream) {
				process.stdout.write(chunk);
				fullText += chunk;
			}
			console.log("\n");

			const assistantMessage: ModelMessage = {
				role: "assistant",
				content: fullText,
			};
			messages.push(assistantMessage);

			const commands = parseTerminalBlocks(fullText);

			if (commands.length === 0) break;

			for (const cmd of commands) {
				const { output, success } = await exec(cmd);

				const resultText = success
					? `COMMAND:\n${cmd}\n\nRESULT:\n${output}`
					: `COMMAND:\n${cmd}\n\nRESULT:\n${output}\n\n[COMMAND FAILED]`;

				// Send result as a separate user message for better context
				const resultMessage: ModelMessage = {
					role: "user",
					content: resultText,
				};
				messages.push(resultMessage);

				if (!success) {
					break;
				}
			}
		}
	}
}
