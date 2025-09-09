# Theme Overrides

This directory is for Hugo layout overrides of the Hextra theme.

## How to Override Theme Files

To override any file from the Hextra theme, create the same file path structure here. For example:

- To override `layouts/_partials/navbar.html`, create `layouts/_partials/navbar.html`  
- To override `layouts/_default/single.html`, create `layouts/_default/single.html`
- To override `layouts/index.html`, create `layouts/index.html`

Hugo will use your local files instead of the theme files when they exist.

## Common Files to Override

- `layouts/_partials/head-meta.html` - Add custom meta tags, analytics, etc.
- `layouts/_partials/navbar.html` - Customize navigation
- `layouts/_partials/footer.html` - Customize footer
- `layouts/_default/single.html` - Customize single page layout
- `layouts/index.html` - Customize home page layout

Any files you create here will be tracked in git and will override the corresponding files in the Hextra theme.
