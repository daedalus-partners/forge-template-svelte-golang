### Daedalus PM – Theming Guidelines (LLM-ready)

This document captures the theme model, design tokens, and usage patterns so another project can adopt the same look-and-feel. It is self-contained and safe to feed into an LLM as guidance.

### Theme model
- **Mode**: dark-only. No toggle; the app always renders in dark.
- **Mechanism**: Design tokens are defined on `:root`. No `data-theme` switching is required.
- **Color scheme hint**: Set `color-scheme: dark` so native form controls match the theme.

### Svelte integration
No theme toggle is required. The app should load with dark tokens and `color-scheme: dark`.

### Design tokens (CSS variables)
Tokens are defined globally and then overridden per theme using `[data-theme='dark']` and `[data-theme='light']` selectors. Copy these into your global CSS (e.g., `index.css`).

```css
/* Dark-only tokens (place on :root) */
:root {
  /* Typography */
  --font-mono: ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', 'DejaVu Sans Mono', 'Ubuntu Mono', 'Courier New', monospace;
  --line-height: 1.5;
  --font-weight-normal: 400;

  /* Spacing scale */
  --space-1: 4px;
  --space-2: 8px;
  --space-3: 12px;
  --space-4: 16px;
  --space-5: 20px;
  --space-6: 24px;

  /* Border radius */
  --radius-sm: 6px;
  --radius-md: 8px;

  /* Colors – dark */
  --color-background: #111112; /* off black */
  --color-surface: #151516;
  --color-elevated: #1B1B1C;
  --color-text: #F2F2F2;
  --color-text-muted: #CCCCCC;
  --color-border: #FFFFFF1A;
  --color-border-muted: #FFFFFF14;
  --color-accent: #FFFFFF;
  --color-accent-strong: #FFFFFF;
  --color-accent-subtle: #222224;

  /* Status */
  --color-success: #22c55e;
  --color-danger: #ef4444;

  /* Aliases */
  --color-bg: var(--color-surface);
  --color-bg-secondary: var(--color-elevated);
  --color-bg-tertiary: var(--color-surface);
  --color-bg-hover: var(--color-accent-subtle);
  --color-text-light: var(--color-text-muted);

  color-scheme: dark;
  color: var(--color-text);
  background-color: var(--color-background);
}
```

### Usage patterns
- **Backgrounds**: Use `var(--color-background)` for the page, `var(--color-surface)` for containers, and `var(--color-elevated)` for cards/dialogs.
- **Text**: Use `var(--color-text)` for primary, `var(--color-text-muted)` for secondary.
- **Borders**: Use `var(--color-border)` and `var(--color-border-muted)` for subtle separators.
- **Accents**: Use `var(--color-accent)` for icons/accents, `var(--color-accent-strong)` for emphasis, and `var(--color-accent-subtle)` for hover or selected states.
- **Status**: Use `--color-success` and `--color-danger` consistently in both themes.
- **Spacing**: Prefer the `--space-*` scale for paddings/margins.
- **Radius**: Use `--radius-sm` and `--radius-md` for corners.

Examples (CSS):

```css
.panel {
  background: var(--color-surface);
  color: var(--color-text);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-md);
  padding: var(--space-4);
}

.button-primary {
  color: var(--color-background);
  background: var(--color-accent);
  border: 1px solid var(--color-border);
  border-radius: var(--radius-sm);
  padding: var(--space-2) var(--space-3);
}

.muted-text { color: var(--color-text-muted); }
```

### Accessibility and UX notes
- Set `color-scheme` to keep native controls aligned with the theme.
- Ensure sufficient contrast when changing token values (check WCAG AA at minimum).
- Avoid hard-coded hex colors in components; consume tokens only.
- Use `outline` styles with `--color-border` for focus-visible states.

### Integration checklist (Svelte app)
- Add dark tokens into your base stylesheet (e.g., `:root`).
- Set `color-scheme: dark`.
- Use tokens throughout components (CSS, inline styles, or utility classes with CSS var values).

### What’s intentionally not included
- No brand color palette beyond neutrals and simple accents; plug in your own brand tokens if needed.
- No Tailwind theme extension required; tokens work with raw CSS or utilities that accept CSS variables.

### Copy-paste summary for LLMs
"Use a dark-only theme. Define CSS variables on `:root` and set `color-scheme: dark`. Use the provided spacing, radius, and status tokens. Do not hardcode colors in components; always consume tokens."


