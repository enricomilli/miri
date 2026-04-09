import hljs from "highlight.js";
import { marked } from "marked";
import { markedHighlight } from "marked-highlight";
import { markedTerminal } from "marked-terminal";

// Configure marked with terminal rendering and syntax highlighting
marked.use(
	markedHighlight({
		langPrefix: "hljs language-",
		highlight(code, lang) {
			const language = hljs.getLanguage(lang) ? lang : "plaintext";
			return hljs.highlight(code, { language }).value;
		},
	}),
);

marked.use(
	markedTerminal({
		width: process.stdout.columns || 80,
		// allowHtml: false,
		// linkFg: "\\x1b[94m",
		reflowText: true,
		tab: 2,
	}) as any,
);

export function renderMarkdown(text: string): string {
	return marked(text) as string;
}

export function renderAndPrint(text: string): void {
	const rendered = renderMarkdown(text);
	process.stdout.write(rendered);
}
