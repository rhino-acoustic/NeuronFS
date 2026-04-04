/**
 * NeuronFS Auto-Injector
 * 
 * THE CORE FEATURE: Watches .neuron files and automatically updates
 * GEMINI.md / .cursorrules / CLAUDE.md with pre-scanned, pre-parsed rules.
 * 
 * This means the AI never needs to scan — rules are already in the system prompt.
 * The extension does the scanning, not the AI.
 * 
 * This is Lv.2 enforcement: the IDE handles the deterministic part (scanning + parsing),
 * and injects the results into the AI's input context.
 */

import * as vscode from 'vscode';
import * as fs from 'fs';
import * as path from 'path';
import { NeuronScanner, RuleFormatter, ScanResult } from './scanner';

// Markers for the injected section
const INJECT_START = '<!-- NEURONFS:START -->';
const INJECT_END = '<!-- NEURONFS:END -->';

export class AutoInjector {
  private scanner: NeuronScanner;
  private watcher: vscode.FileSystemWatcher | undefined;
  private debounceTimer: NodeJS.Timeout | undefined;
  private statusBar: vscode.StatusBarItem;

  constructor(statusBar: vscode.StatusBarItem) {
    this.scanner = new NeuronScanner();
    this.statusBar = statusBar;
  }

  /**
   * Start watching .neuron files and auto-inject on changes
   */
  startWatching(): void {
    // Watch for .neuron file changes across the workspace
    this.watcher = vscode.workspace.createFileSystemWatcher('**/*.neuron');

    this.watcher.onDidCreate(() => this.debouncedInject('created'));
    this.watcher.onDidDelete(() => this.debouncedInject('deleted'));
    this.watcher.onDidChange(() => this.debouncedInject('changed'));

    // Also watch global neurons path
    const config = vscode.workspace.getConfiguration('neuronfs');
    const globalPath = config.get<string>('globalNeuronsPath');
    if (globalPath) {
      const globalPattern = new vscode.RelativePattern(globalPath, '**/*.neuron');
      const globalWatcher = vscode.workspace.createFileSystemWatcher(globalPattern);
      globalWatcher.onDidCreate(() => this.debouncedInject('created'));
      globalWatcher.onDidDelete(() => this.debouncedInject('deleted'));
      globalWatcher.onDidChange(() => this.debouncedInject('changed'));
    }
  }

  /**
   * Debounce rapid changes (e.g., multiple neurons created at once)
   */
  private debouncedInject(reason: string): void {
    if (this.debounceTimer) clearTimeout(this.debounceTimer);
    this.debounceTimer = setTimeout(async () => {
      await this.inject();
      vscode.window.showInformationMessage(`🧠 NeuronFS: Rules auto-updated (${reason})`);
    }, 500);
  }

  /**
   * Scan neurons and inject into the target system prompt file
   */
  async inject(): Promise<ScanResult> {
    const result = await this.scanner.scan();
    const config = vscode.workspace.getConfiguration('neuronfs');
    const target = config.get<string>('injectTarget') || 'disabled.md';

    // Find all target files to inject into
    const targetPaths = await this.findTargetFiles(target);

    for (const targetPath of targetPaths) {
      await this.injectIntoFile(targetPath, result);
    }

    // Update status bar
    const activeCount = result.neurons.filter(n => !n.isDormant).length;
    this.statusBar.text = `$(brain) ${activeCount} neurons`;
    this.statusBar.tooltip = `NeuronFS: ${activeCount} active, ${result.totalWeight}B total weight`;

    return result;
  }

  /**
   * Find target files for injection
   */
  private async findTargetFiles(target: string): Promise<string[]> {
    const paths: string[] = [];

    if (!vscode.workspace.workspaceFolders) return paths;

    for (const folder of vscode.workspace.workspaceFolders) {
      const root = folder.uri.fsPath;

      switch (target) {
        case 'gemini.md': {
          // Check .gemini/GEMINI.md (Antigravity)
          const geminiPath = path.join(root, '.gemini', 'GEMINI.md');
          if (fs.existsSync(geminiPath)) paths.push(geminiPath);
          // Also check home directory
          const homeGemini = path.join(
            process.env.USERPROFILE || process.env.HOME || '',
            '.gemini', 'GEMINI.md'
          );
          if (fs.existsSync(homeGemini) && !paths.includes(homeGemini)) {
            paths.push(homeGemini);
          }
          break;
        }
        case '.cursorrules': {
          const cursorPath = path.join(root, '.cursorrules');
          paths.push(cursorPath); // Create if doesn't exist
          break;
        }
        case 'claude.md': {
          const claudePath = path.join(root, 'CLAUDE.md');
          paths.push(claudePath);
          break;
        }
      }
    }

    return paths;
  }

  /**
   * Inject formatted rules into a specific file
   * 
   * Strategy:
   * - If file has NEURONFS:START / NEURONFS:END markers → replace between them
   * - If file exists but no markers → append markers + rules at the end
   * - If file doesn't exist → create with rules
   */
  private async injectIntoFile(filePath: string, result: ScanResult): Promise<void> {
    const formatted = RuleFormatter.formatForInjection(result);
    const injectionBlock = `${INJECT_START}\n${formatted}\n${INJECT_END}`;

    try {
      if (fs.existsSync(filePath)) {
        let content = fs.readFileSync(filePath, 'utf-8');
        const startIdx = content.indexOf(INJECT_START);
        const endIdx = content.indexOf(INJECT_END);

        if (startIdx !== -1 && endIdx !== -1) {
          // Replace existing injection block (wherever it is)
          content = content.substring(0, startIdx) +
            injectionBlock +
            content.substring(endIdx + INJECT_END.length);
        } else {
          // PREPEND at TOP — rules must be the first thing the AI reads
          content = injectionBlock + '\n\n' + content;
        }

        fs.writeFileSync(filePath, content, 'utf-8');
      } else {
        // Create file with injection block
        const dir = path.dirname(filePath);
        if (!fs.existsSync(dir)) fs.mkdirSync(dir, { recursive: true });
        fs.writeFileSync(filePath, injectionBlock + '\n', 'utf-8');
      }
    } catch (err) {
      vscode.window.showErrorMessage(`NeuronFS: Failed to inject into ${filePath}: ${err}`);
    }
  }

  dispose(): void {
    if (this.watcher) this.watcher.dispose();
    if (this.debounceTimer) clearTimeout(this.debounceTimer);
  }
}
