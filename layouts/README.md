# Adobe Security Digest - Hugo Layouts

Custom Hugo layouts and templates for the Adobe Security Digest website.

---

## 📐 Layout Architecture

This directory contains custom Hugo layouts that provide the professional, branded experience for Adobe Security Digest. The layouts are designed to be:

- **🎨 Responsive** - Mobile-first design with clean typography
- **⚡ Fast** - Minimal CSS and optimized for performance  
- **♿ Accessible** - Semantic HTML with proper ARIA labels
- **🔍 SEO-Optimized** - Structured data and meta tags

---

## 🗂️ Layout Structure

### Core Templates

| File | Purpose | Features |
|------|---------|----------|
| `index.html` | Homepage template | Adobe branding, statistics, navigation |
| `products/list.html` | Products overview page | Card grid, RSS feed links |
| `products/single.html` | Individual product pages | Bulletin lists, product-specific RSS |

### Template Features

#### `index.html` - Homepage
- **Adobe Brand Colors** - Professional red gradient background
- **Dynamic Statistics** - Bulletin counts and last update info
- **Credit Attribution** - Links to Doug Hatcher and Blue Acorn iCi
- **Responsive Design** - Mobile-optimized layout
- **Call-to-Action** - Clear navigation to bulletins and products

#### `products/list.html` - Products Overview  
- **Product Card Grid** - Clean, organized product listing
- **Bulletin Counts** - Shows number of advisories per product
- **RSS Feed Integration** - Links to product-specific RSS feeds
- **Search Guidance** - Helper text for users
- **Professional Styling** - Consistent with Adobe branding

#### `products/single.html` - Product Pages
- **Breadcrumb Navigation** - Easy navigation hierarchy
- **RSS Feed Access** - Direct links to product RSS feeds  
- **Clean Typography** - Readable bulletin information
- **Action Buttons** - Navigation to related pages

---

## 🎨 Design System

### Color Palette
```css
/* Adobe Brand Colors */
--adobe-red: #FF0000;
--adobe-red-dark: #CC0000; 
--adobe-red-darker: #800000;

/* Supporting Colors */
--text-primary: #942a2a;
--text-secondary: #6b7280;
--background: #ffffff;
--border: #e5e7eb;
```

### Typography
- **Font Stack**: `-apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Oxygen, Ubuntu, Cantarell, sans-serif`
- **Headings**: Bold, high contrast for accessibility
- **Body Text**: Optimized line height (1.6-1.7) for readability
- **Small Text**: Used for metadata and secondary information

### Layout Principles
- **Mobile First** - Responsive design starts with mobile
- **Grid Systems** - CSS Grid for complex layouts, Flexbox for components
- **Consistent Spacing** - Systematic use of padding and margins
- **Visual Hierarchy** - Clear information architecture

---

## 🔧 Customization

### Adding New Templates

1. **Create the layout file** in the appropriate directory:
   ```
   layouts/
   ├── section-name/
   │   ├── list.html      # Section overview page
   │   └── single.html    # Individual item page
   ```

2. **Follow the design system**:
   - Use consistent CSS classes and styling
   - Include proper responsive breakpoints
   - Add RSS feed integration where appropriate

3. **Test thoroughly**:
   ```bash
   hugo server --bind 0.0.0.0 --port 1313
   ```

### Style Guidelines

#### CSS Organization
- **Inline Styles** - Used for component-specific styling
- **Consistent Naming** - Descriptive class names (e.g., `product-card`, `bulletin-meta`)
- **Responsive Design** - Media queries for mobile optimization
- **Performance** - Minimal CSS, avoid heavy frameworks

#### HTML Structure
- **Semantic HTML** - Proper use of headers, nav, main, sections
- **Accessibility** - ARIA labels, alt text, keyboard navigation
- **SEO Optimization** - Structured data, meta tags, proper headings
- **Clean Markup** - Well-indented, commented code

---

## 📱 Responsive Breakpoints

```css
/* Mobile devices */
@media (max-width: 768px) {
  /* Single column layouts */
  /* Larger touch targets */
  /* Simplified navigation */
}

/* Tablet and up */
@media (min-width: 769px) {
  /* Multi-column grids */
  /* Hover effects */
  /* Extended navigation */
}
```

---

## 🚀 Performance Optimization

### Loading Performance
- **Minimal CSS** - Only necessary styles, no external frameworks
- **Optimized Images** - Proper sizing and formats
- **Clean HTML** - Semantic markup without bloat

### Runtime Performance  
- **CSS Grid/Flexbox** - Modern layout techniques
- **Efficient Selectors** - Avoid complex CSS selectors
- **Minimal JavaScript** - Static site with minimal client-side code

---

## 🔍 SEO & Accessibility

### Search Engine Optimization
- **Structured Data** - JSON-LD for rich snippets
- **Meta Tags** - Proper title, description, and Open Graph tags
- **Clean URLs** - Human-readable, hierarchical structure
- **XML Sitemaps** - Automatic generation via Hugo

### Accessibility Features
- **Semantic HTML** - Proper heading hierarchy and landmarks
- **Color Contrast** - WCAG AA compliance
- **Keyboard Navigation** - Full keyboard accessibility
- **Screen Reader Support** - ARIA labels and descriptions

---

## 🛠️ Development Workflow

### Local Development
```bash
# Start Hugo development server
hugo server --bind 0.0.0.0 --port 1313

# Watch for layout changes (automatic reload)
# Edit templates in layouts/ directory
# Changes appear immediately in browser
```

### Testing Layouts
```bash
# Build static site
hugo --minify

# Validate HTML structure
# Test responsive design
# Check RSS feed generation
```

### Production Deployment
- **Automated Builds** - GitHub Actions builds and deploys
- **Cache Optimization** - Proper cache headers for static assets
- **CDN Distribution** - Fast global content delivery

---

**🎨 Professional Design**: Clean, accessible, and optimized for Adobe Security Digest's professional use cases.
