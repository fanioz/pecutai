/**
 * Loading verbs, in the spirit of the Claude Code spinner's rotating gerunds
 * ("Cogitating…", "Percolating…"). Kept short so they fit a typing bubble,
 * and on-brand with the WUPHF voice (honest, funny, AI workplace).
 *
 * Microcopy follows the house style: no contractions, no em-dashes. They cycle
 * decoratively, so they are aria-hidden behind a stable status label.
 */
export const OFFICE_LOADING_PHRASES = [
  "Reticulating splines",
  "Computing",
  "Optimizing",
  "Loading",
  "Processing",
  "Calculating",
  "Indexing",
  "Compiling",
  "Synthesizing",
  "Parsing",
  "Generating",
  "Fetching",
  "Updating",
  "Rendering",
  "Executing",
  "Evaluating",
  "Computing embeddings",
  "Running inference",
  "Tuning hyperparameters",
  "Checking dependencies",
] as const;
