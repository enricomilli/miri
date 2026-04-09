export function parseTerminalBlocks(text: string): string[] {
	const blocks: string[] = [];
	const regex = /```terminal\n([\s\S]*?)```/g;
	let match: RegExpExecArray | null;
	while (true) {
		match = regex.exec(text);
		if (match === null) break;
		const block = match[1];
		if (block !== undefined) {
			blocks.push(block.trim());
		}
	}
	return blocks;
}
