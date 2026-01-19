const esbuild = require("esbuild");
const fs = require("fs");
const path = require("path");

const production = process.argv.includes('--production');
const watch = process.argv.includes('--watch');

/**
 * @type {import('esbuild').Plugin}
 */
const esbuildProblemMatcherPlugin = {
	name: 'esbuild-problem-matcher',

	setup(build) {
		build.onStart(() => {
			console.log('[watch] build started');
		});
		build.onEnd((result) => {
			result.errors.forEach(({ text, location }) => {
				console.error(`✘ [ERROR] ${text}`);
				console.error(`    ${location.file}:${location.line}:${location.column}:`);
			});
			console.log('[watch] build finished');
		});
	},
};

async function main() {
	// Build extension
	const extensionCtx = await esbuild.context({
		entryPoints: ['src/extension.ts'],
		bundle: true,
		format: 'cjs',
		minify: production,
		sourcemap: !production,
		sourcesContent: false,
		platform: 'node',
		outfile: 'dist/extension.js',
		external: ['vscode'],
		logLevel: 'silent',
		plugins: [esbuildProblemMatcherPlugin],
	});

	// Build webview
	const webviewCtx = await esbuild.context({
		entryPoints: ['src/webviews/chat/main.ts'],
		bundle: true,
		format: 'iife',
		minify: production,
		sourcemap: !production,
		sourcesContent: false,
		platform: 'browser',
		outfile: 'dist/webviews/chat/main.js',
		logLevel: 'silent',
		plugins: [esbuildProblemMatcherPlugin],
	});

	// Copy CSS files
	const copyCss = () => {
		const cssSource = path.join(__dirname, 'src', 'webviews', 'chat', 'styles.css');
		const cssDest = path.join(__dirname, 'dist', 'webviews', 'chat', 'styles.css');
		const destDir = path.dirname(cssDest);
		if (!fs.existsSync(destDir)) {
			fs.mkdirSync(destDir, { recursive: true });
		}
		fs.copyFileSync(cssSource, cssDest);
	};

	if (watch) {
		copyCss();
		// Watch for CSS changes
		fs.watchFile(cssSource, () => {
			copyCss();
			console.log('[watch] CSS copied');
		});
		await Promise.all([extensionCtx.watch(), webviewCtx.watch()]);
	} else {
		copyCss();
		await Promise.all([
			extensionCtx.rebuild(),
			webviewCtx.rebuild(),
		]);
		await Promise.all([extensionCtx.dispose(), webviewCtx.dispose()]);
	}
}

main().catch(e => {
	console.error(e);
	process.exit(1);
});
