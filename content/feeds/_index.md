---
title: "RSS Feeds"
description: "Subscribe to Adobe security bulletin RSS feeds for real-time updates"
---

# RSS Feeds

Stay updated with the latest Adobe security bulletins through our RSS feeds. Choose from comprehensive feeds covering all products or product-specific feeds for targeted alerts.

## 📡 Available Feeds

{{< cards >}}
  {{< card link="/feeds/bulletins.xml" title="🛡️ All Security Bulletins" subtitle="Complete feed with all Adobe security updates" >}}
  {{< card link="/feeds/commerce-magento.xml" title="🛒 Adobe Commerce" subtitle="Commerce and Magento security bulletins" >}}
  {{< card link="/feeds/experience-manager.xml" title="📄 Experience Manager" subtitle="AEM security bulletins and updates" >}}
{{< /cards >}}

## 🔔 How to Subscribe

### Desktop Feed Readers
- **Feedly**: Copy the feed URL and paste it into Feedly
- **Inoreader**: Add subscription using the feed URL
- **NewsBlur**: Subscribe using the RSS feed link

### Mobile Apps
- **iOS**: Use apps like NetNewsWire, Reeder, or Feedly mobile app
- **Android**: Try Feedly, Inoreader, or FeedReader

### Email Notifications
Some RSS services like Feedly and IFTTT can convert RSS feeds to email notifications.

## 📊 Feed Information

{{< callout emoji="⏱️" >}}
**Update Frequency**  
All feeds are updated every 6 hours to ensure you receive timely security notifications.
{{< /callout >}}

{{< callout emoji="📋" >}}
**Feed Content**  
Each feed item includes:
- Security bulletin title and ID
- Publication and update dates  
- Priority level (Critical, Important, Moderate, Low)
- CVE identifiers and CVSS scores
- Affected product versions
- Link to full bulletin details
{{< /callout >}}

## 🛠️ Technical Details

- **Format**: RSS 2.0 standard
- **Encoding**: UTF-8
- **Maximum items**: 50 most recent bulletins per feed
- **Update schedule**: Every 6 hours via automated GitHub Actions

---

{{< callout type="info" >}}
**Need Help?**

Having trouble with RSS feeds? Check your feed reader's documentation or contact us through our [About page](/about/).
{{< /callout >}}
