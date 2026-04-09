import { $ } from "bun";

export interface ExecResult {
	output: string;
	success: boolean;
}

export async function exec(cmd: string): Promise<ExecResult> {
	console.log(`\x1b[36m  → Running: ${cmd}\x1b[0m`);

	try {
		const proc = await $`${{ raw: cmd }}`.quiet().nothrow();
		const stdout = proc.stdout.toString().trim();
		const stderr = proc.stderr.toString().trim();
		const code = proc.exitCode;

		let output = "";
		if (stdout) output += stdout;
		if (stderr) output += (output ? "\n" : "") + `[stderr] ${stderr}`;

		const success = code === 0;
		if (success) {
			console.log(`\x1b[32m  ✓ Success\x1b[0m`);
		} else {
			output += (output ? "\n" : "") + `[exit code ${code}]`;
			console.log(`\x1b[31m  ✘ Failed (exit code ${code})\x1b[0m`);
		}

		return {
			output: output || "(no output)",
			success,
		};
	} catch (err) {
		const errorMessage = err instanceof Error ? err.message : String(err);
		console.log(`\x1b[31m  ✘ Error: ${errorMessage}\x1b[0m`);
		return {
			output: `[execution error] ${errorMessage}`,
			success: false,
		};
	}
}
