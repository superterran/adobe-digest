---
title: "About Adobe Digest"
description: "Learn about our mission to keep Adobe users informed about security updates"
---

# About Adobe Digest

## 🛡️ Our Mission

Adobe Digest is an **unofficial** security intelligence platform dedicated to helping Adobe users stay informed about critical security updates. We automatically monitor and organize security bulletins from Adobe's Product Security Incident Response Team (PSIRT) to ensure you never miss an important security update.

## 🎯 What We Do

{{< cards >}}
  {{< card title="🔍 Automated Monitoring" subtitle="24/7 tracking of Adobe security advisories" >}}
  {{< card title="📊 Intelligent Organization" subtitle="Structured, searchable security bulletin database" >}}
  {{< card title="⚡ Real-time Delivery" subtitle="RSS feeds and web updates every 6 hours" >}}
  {{< card title="🎨 Clean Interface" subtitle="Easy-to-read security information" >}}
{{< /cards >}}

## 🏗️ Products Covered

- **Adobe Commerce & Magento** - E-commerce platform security updates
- **Adobe Experience Manager (AEM)** - Content management system vulnerabilities  
- **Adobe Experience Platform** - Customer data platform security advisories

## 🔧 How It Works

{{< callout emoji="🤖" >}}
**Automated Pipeline**

1. **Monitor**: Our system checks Adobe's security pages every 6 hours
2. **Parse**: Extract and structure security bulletin information
3. **Organize**: Create searchable, categorized content
4. **Deliver**: Generate RSS feeds and update the website
5. **Deploy**: Automatically publish updates via GitHub Actions
{{< /callout >}}

## 📊 Technical Architecture

- **Backend**: Go-based scraper with HTML parsing and RSS generation
- **Frontend**: Hugo static site generator with Hextra theme  
- **Hosting**: GitHub Pages with custom domain
- **Automation**: GitHub Actions for scheduling and deployment
- **Data Sources**: Adobe's official PSIRT security bulletins

## ⚠️ Important Disclaimers

{{< callout type="warning" >}}
**Unofficial Service**

Adobe Digest is **not affiliated with Adobe Inc.** We are an independent service that aggregates publicly available security information. Always refer to [Adobe's official security advisories](https://helpx.adobe.com/security.html) for authoritative information.
{{< /callout >}}

{{< callout type="info" >}}
**Data Accuracy**

While we strive for accuracy, security information is automatically processed and may contain errors. Always verify critical security information against Adobe's official sources before taking action.
{{< /callout >}}

## 🤝 Open Source

This project is open source and available on GitHub. We welcome contributions, bug reports, and feature requests.

- **Repository**: [github.com/superterran/adobe-digest](https://github.com/superterran/adobe-digest)
- **License**: MIT License
- **Contributions**: Pull requests welcome

## 📞 Contact & Support

- **Issues**: Report bugs via GitHub Issues
- **Feature Requests**: Submit via GitHub Discussions  
- **General Questions**: Create a GitHub Issue with the "question" label

## 🔄 Update Schedule

- **Security Bulletins**: Checked every 6 hours
- **Website Updates**: Automatic deployment on content changes
- **RSS Feeds**: Updated with each scraper run
- **System Status**: Monitored via GitHub Actions workflow status

---

{{< callout emoji="💡" >}}
**Pro Tip**

Subscribe to our RSS feeds for the fastest notifications of new security bulletins. RSS readers can often deliver alerts faster than checking websites manually.
{{< /callout >}}

## 🏆 Credits

Built with:
- [Hugo](https://gohugo.io/) - Static site generator
- [Hextra](https://github.com/imfing/hextra) - Documentation theme
- [GitHub Actions](https://github.com/features/actions) - CI/CD automation
- [GitHub Pages](https://pages.github.com/) - Hosting platform

Special thanks to Adobe's Product Security Incident Response Team for maintaining comprehensive public security advisories.
