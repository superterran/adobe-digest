# Adobe Digest

A Hugo site using the [Hextra](https://github.com/imfing/hextra) theme, built with minimal footprint using Hugo modules.

## 🚀 Live Site

The site is automatically deployed to GitHub Pages: https://superterran.github.io/adobe-digest

## 🏗️ Development

### Prerequisites

- Hugo (extended version)
- Go (for module management)

### Local Development

1. Clone the repository:
   ```bash
   git clone https://github.com/superterran/adobe-digest.git
   cd adobe-digest
   ```

2. Get Hugo modules:
   ```bash
   hugo mod get
   ```

3. Start the development server:
   ```bash
   hugo server --buildDrafts
   ```

4. Open your browser to `http://localhost:1313/adobe-digest/`

### Building

To build the site for production:

```bash
hugo --gc --minify
```

## 🎨 Theme Customization

This project uses the Hextra theme as a Hugo module, keeping the repository lightweight.

### Making Theme Overrides

To customize the theme:

1. Create files in the `layouts/` directory matching the theme's structure
2. Hugo will automatically use your local files instead of the theme files
3. Your overrides will be tracked in git while the base theme remains external

See `layouts/README.md` for detailed instructions.

### Updating the Theme

To update to the latest version of Hextra:

```bash
hugo mod get -u
```

## 📁 Project Structure

```
├── content/           # Site content (Markdown files)
├── layouts/           # Theme overrides and custom layouts
├── static/            # Static assets
├── .github/workflows/ # GitHub Actions for deployment
├── hugo.toml         # Hugo configuration
├── go.mod            # Go module dependencies (theme management)
└── go.sum            # Go module checksums
```

## 🚢 Deployment

The site is automatically deployed to GitHub Pages using GitHub Actions whenever changes are pushed to the `main` branch.

The deployment workflow:
1. Sets up Hugo and Go
2. Downloads theme dependencies via Hugo modules
3. Builds the site
4. Deploys to GitHub Pages

## 📝 Adding Content

Create new content files in the `content/` directory:

```bash
hugo new posts/my-new-post.md
```

Edit the generated file and start writing in Markdown!
