## My coding style (TypeScript)

These rules apply whatever the framework. When a framework requires a specific form (file names, exports, component syntax, routing folders), the framework wins. The framework sections at the end cover the main cases.

### Language
- Identifiers, comments, commit messages and PR descriptions: English. User-facing text: French.
  Do: `const monthlyRent = …` · Don't: `const loyerMensuel = …`
- Never add an AI co-author trailer or a "generated with" footer to commits or PRs.

### Formatting (the project's formatter wins on anything it decides)
- 2 spaces, single quotes, semicolons, trailing commas, 120-column lines.
- Braces on every `if`, even one-liners.
  Do: `if (!user) {\n  return null;\n}` · Don't: `if (!user) return null;`
- Leave a blank line before `return` when other statements precede it.
- Avoid `else`. Return or continue instead.
- The framework's imports come first (`react`, `astro:content`, `vue`…). Type-only imports use separate `import type { … }` statements.

### Naming
- Components: one per file, PascalCase, with the framework's extension (`Card.tsx`, `Card.astro`, `Card.vue`). Hooks and composables: `useX.ts`. Other modules: camelCase.
- Top-level folders are lowercase plurals grouped by kind (`components/`, `hooks/`, `helpers/`, `services/`, `data/`). Folders the framework requires (`pages/`, `app/`, `routes/`, `layouts/`) keep the framework's names.
- Props types are named `<Component>Props`, declared in the component's file, and not exported unless another file needs them. Exception: use the name the framework expects (Astro: `Props`).
- Internal handlers are `handleX`; callback props are `onX`; the event parameter is `event`, never `e`.
  Do: `const handleSubmit = (event: SubmitEvent) => …` · Don't: `const onSubmit = (e) => …`
- Booleans start with `is`/`has`/`should`. Do: `isOpen` · Don't: `open`
- Use full words. Abbreviations only in one-line callbacks.

### Functions & components
- Arrow functions only. Never `function`, never `class` unless the code genuinely needs inheritance.
- Destructure props where they arrive: in the parameter list, or right after the framework's props API.
- Order inside a component, whatever the syntax: constants and refs → state → other hooks or data loading → effects → derived values → handlers → markup.
- In `.ts`/`.tsx` modules, put all exports at the end of the file: `export type { … };` then `export { … };`. Never inline `export const`. Use a default export only where the framework requires one. Exports the framework requires inside a single-file component (`.astro`, `.vue`, `.svelte`) go where the framework expects them.
- Use `.map`, `.filter`, `.reduce` or `for…of`; never `.forEach`. Use `const`; `let` only when reassigned.
- `async`/`await`, not `.then()` chains.

### Size and complexity
Apply these limits to code you write or change. Don't refactor untouched code just to meet them.
- A function is 40 lines at most. A component is 150 lines at most, markup included; beyond that, split out child components.
- A file is 300 non-blank lines at most. Split it by responsibility.
- Cognitive complexity is 15 at most per function.
- At most 3 nested levels of `if`/`for`/`while`/`switch`/`try`. Flatten with early returns or extract a function.
- At most 3 parameters. Beyond that, take one object and destructure it.
- No nested ternaries. Use early returns or a lookup object.
- Give numbers whose meaning isn't obvious (delays, limits, sizes) a named constant.
- No commented-out code: delete it, git keeps the history.
- No unused variables, imports, parameters or assignments.
- Don't duplicate logic. Import the existing constant or helper; when the same logic appears twice, extract it once.

### Types
- `type`, never `interface` (except global augmentation).
- Strict mode. No `any`, no non-null `!`, no `enum` (use string unions). Avoid `as` casts.
- Keep types next to their use. Shared types come from their source (schema, helper), not a central `types/` file.
- Type DOM element references with the exact element type: `document.querySelector<HTMLDialogElement>('#menu')`, `useRef<HTMLDivElement | null>(null)` in React.

### Errors
- Trust internal code. Validate only at boundaries (request bodies, external APIs, env).
- No custom error classes. Write `catch {` without a binding when the error isn't used.
- No `console.log` in committed code. Use the project's logger.

### Comments
- Default to no comment. When needed: English, one line (two at most), explaining *why*, not *what*.
  Do: `// Delete sessions first to avoid FK constraint errors` · Don't: `// Delete the user`
- No TODO/FIXME, no JSDoc paragraphs, no `@param` blocks.

### Dependencies & UI
- Write small utilities and hooks yourself before adding a dependency. Icons come from Lucide, using the package for the project's framework (`lucide-react` in React).
- Style with Tailwind. Use plain HTML elements with classes; reach for a component library only when the project already does.
- Animations: CSS `@keyframes` in the Tailwind theme; no motion library.
- Compose classes with the project's helper (`cn()`) or the framework's built-in (`class:list` in Astro) if there is one, otherwise template literals. Never mix two approaches in one file.
- French copy: a space before `?`, `!` and `:`; straight apostrophes; address the reader as "vous".
  Do: `Prêt à démarrer ?` · Don't: `Prêt à démarrer?`
- `aria-label` in French; decorative icons get `aria-hidden="true"`; `target="_blank"` always comes with `rel="noopener noreferrer"`.
- Obfuscate email addresses and phone numbers in the markup.

### Architecture
- Few layers; no frameworks or patterns for their own sake. Keep pure helpers separate from I/O.
- Put static content in typed data modules, not inline in components.
- No barrel `index.ts` files.

### Git
- Conventional Commits, lowercase, short subject: `feat: add rentability calculator`, `fix: typo`, `chore: update dependencies`.

### React (when the project uses it)
- Type every component with `FC`, including async server components; destructure props in the parameter list.
  Do: `const Card: FC<CardProps> = ({ title }) => (` / `const Page: FC = async () => {`
  Don't: `export default function Card(props: CardProps) {`
- A component that only returns JSX uses a concise body `=> (`.
- Never use an array index as a list `key`; use a stable id from the data.

### Astro (when the project uses it)
- Props: `type Props = { … };` in the frontmatter (Astro reads the name `Props`), not the `interface Props` the Astro docs show. Destructure right after: `const { title, isActive = false } = Astro.props;`.
- Frontmatter order: imports → `type Props` → `Astro.props` → data loading → derived values.
- Exports Astro requires stay inline, written as arrow functions: `export const getStaticPaths = (async () => { … }) satisfies GetStaticPaths;`.
- For interactivity, write a typed `<script>` in the component before adding a UI framework island.

### Other frameworks (Vue, Svelte, Solid…)
- Use the framework's required syntax (props API, file names, routing folders). Apply the rules above to everything the framework leaves open.
- Declare props as a `type` (never `interface`) and pass it to the framework's props API.
