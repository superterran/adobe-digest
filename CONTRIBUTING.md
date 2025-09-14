# Contributing to Adobe Security Digest

Thank you for your interest in contributing to Adobe Security Digest! This document provides guidelines and information for contributors.

## Code of Conduct

This project follows a standard code of conduct:
- Be respectful and inclusive
- Focus on constructive feedback
- Help others learn and grow
- Maintain a professional tone in all interactions

## How to Contribute

### 🐛 Reporting Bugs

Before creating bug reports, please:
1. Check existing issues to avoid duplicates
2. Use the issue template if available
3. Provide clear, detailed information including:
   - Steps to reproduce
   - Expected vs actual behavior
   - Environment details (OS, Go version, etc.)
   - Relevant logs or error messages

### 💡 Suggesting Enhancements

Enhancement suggestions are welcome! Please:
1. Check existing feature requests first
2. Provide a clear description of the proposed feature
3. Explain the use case and benefits
4. Consider implementation complexity and maintenance burden

### 🔧 Code Contributions

#### Development Setup

1. **Fork and clone the repository**:
   ```bash
   git clone https://github.com/your-username/adobe-digest.git
   cd adobe-digest
   ```

2. **Install dependencies**:
   ```bash
   go mod download
   ```

3. **Set up development environment**:
   ```bash
   # Install Hugo (if not already installed)
   # Run scraper to populate data
   just scrape
   # Start development server
   just dev
   ```

#### Making Changes

1. **Create a feature branch**:
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**:
   - Write clean, readable code
   - Follow Go conventions and best practices
   - Add comments for complex logic
   - Update documentation if needed

3. **Test your changes**:
   ```bash
   # Test scraper functionality
   go run cmd/adobe-scraper/main.go test
   
   # Test content generation
   go run cmd/content-generator/main.go generate
   
   # Build Hugo site
   hugo --minify
   ```

4. **Commit your changes**:
   ```bash
   git add .
   git commit -m "feat: add new feature description"
   ```
   
   Use conventional commit messages:
   - `feat:` for new features
   - `fix:` for bug fixes
   - `docs:` for documentation changes
   - `refactor:` for code refactoring
   - `test:` for adding tests

5. **Push and create pull request**:
   ```bash
   git push origin feature/your-feature-name
   ```

#### Pull Request Guidelines

- **Clear title and description**: Explain what changes you made and why
- **Link related issues**: Reference any related issue numbers
- **Small, focused changes**: Keep PRs focused on a single feature or fix
- **Update documentation**: Include documentation updates for user-facing changes
- **Test thoroughly**: Ensure all functionality works as expected

### 🏗️ Architecture Guidelines

When contributing code, please consider:

#### Scraper (`cmd/adobe-scraper/`)
- Maintain the multi-strategy approach
- Add new strategies as separate functions
- Handle errors gracefully with informative messages
- Test against live Adobe endpoints carefully

#### Content Generator (`cmd/content-generator/`)
- Keep Hugo template generation flexible
- Maintain RSS feed compatibility
- Ensure data integrity in bulletin database
- Handle edge cases in bulletin data

#### Documentation
- Keep README files current with code changes
- Use clear, professional language
- Include examples where helpful
- Maintain consistency in formatting

## Development Philosophy

### Reliability First
- Adobe's infrastructure changes frequently
- Build robust error handling and fallback mechanisms
- Test thoroughly before releasing changes

### User Experience
- Prioritize clean, fast RSS feeds
- Maintain consistent data formats
- Keep the website lightweight and accessible

### Automation
- Minimize manual intervention requirements
- Design for GitHub Actions compatibility
- Consider long-term maintenance burden

## Getting Help

If you need help with contributing:

1. **Check existing documentation** in README files
2. **Search existing issues** for similar questions
3. **Create a discussion** for general questions
4. **Open an issue** for specific problems

## Recognition

Contributors will be recognized in:
- Repository contributor list
- Release notes for significant contributions
- Special thanks for major features or fixes

## Questions?

Feel free to reach out by:
- Opening a GitHub issue
- Starting a GitHub discussion
- Contacting [@doughatcher](https://github.com/doughatcher)

Thank you for helping make Adobe Security Digest better! 🚀
