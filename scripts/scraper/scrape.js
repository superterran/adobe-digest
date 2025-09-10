// Minimal, modular scraping harness for Adobe Digest
// - Fetches security bulletins for Adobe AEM, Adobe Commerce (Magento), and AEP (placeholder)
// - Writes Markdown into content/bulletins/{vendor}/{year}/{slug}.md with front matter
// - Idempotent based on a computed stable id

import { mkdir, writeFile, readFile } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import crypto from 'node:crypto';
import cheerio from 'cheerio';
import { fetch } from 'undici';

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

// Config
const ROOT = path.resolve(__dirname, '../../');
const CONTENT_DIR = path.join(ROOT, 'content', 'bulletins');

// Utilities
function hashId(str) {
  return crypto.createHash('sha256').update(str).digest('hex').slice(0, 12);
}

function toSlug(s) {
  return s
    .toLowerCase()
    .replace(/\s+/g, '-').replace(/[^a-z0-9\-]/g, '')
    .replace(/-+/g, '-')
    .replace(/^-|-$|_/g, '');
}

function frontMatter(obj) {
  // Very small YAML serializer for simple scalars/arrays
  const lines = Object.entries(obj).map(([k, v]) => {
    if (Array.isArray(v)) return `${k}: [${v.map(x => JSON.stringify(x)).join(', ')}]`;
    if (typeof v === 'string') return `${k}: ${JSON.stringify(v)}`;
    if (v === null || v === undefined) return `${k}:`;
    return `${k}: ${v}`;
  });
  return `---\n${lines.join('\n')}\n---\n`;
}

async function writeMarkdown({ vendor, product, title, date, severity, cves = [], url, body = '' }) {
  const year = new Date(date).getFullYear();
  const baseSlug = toSlug(title || `${vendor}-${product}-${date}`);
  const id = hashId(`${vendor}|${product}|${url}`);
  const slug = `${baseSlug}-${id}`;
  const dir = path.join(CONTENT_DIR, vendor.toLowerCase(), String(year));
  const filePath = path.join(dir, `${slug}.md`);

  if (!existsSync(dir)) await mkdir(dir, { recursive: true });

  // Idempotency: if file exists and contains the same source URL id, skip
  if (existsSync(filePath)) {
    try {
      const existing = await readFile(filePath, 'utf8');
      if (existing.includes(`source_url: \"${url}\"`) || existing.includes(url)) {
        console.log(`[skip] ${title} already exists: ${filePath}`);
        return { skipped: true, filePath };
      }
    } catch {}
  }

  const fm = frontMatter({
    title,
    date,
    vendor,
    product,
    severity: severity || 'unknown',
    cves,
    source_url: url,
    draft: false,
    tags: [vendor, product, 'security-bulletin']
  });

  const content = `${fm}\n${body}\n`;
  await writeFile(filePath, content, 'utf8');
  console.log(`[write] ${filePath}`);
  return { skipped: false, filePath };
}

// Scrapers
async function scrapeAdobeCommerce() {
  // Example page given: APSB25-88 Magento/Adobe Commerce
  // Root listing pages are stable: https://helpx.adobe.com/security.html -> Products -> Commerce
  const listUrl = 'https://helpx.adobe.com/security/products/magento.html';
  const res = await fetch(listUrl, { headers: { 'user-agent': 'adobe-digest-bot/0.1 (+https://adobedigest.com)' } });
  if (!res.ok) throw new Error(`Failed to fetch ${listUrl}: ${res.status}`);
  const html = await res.text();
  const $ = cheerio.load(html);

  const links = [];
  $('a').each((_, a) => {
    const href = $(a).attr('href') || '';
    const text = $(a).text().trim();
    if (/\/security\/products\/magento\//.test(href) && /APSB/i.test(text)) {
      const url = new URL(href, listUrl).toString();
      links.push({ url, text });
    }
  });

  const results = [];
  for (const { url } of links.slice(0, 30)) {
    try {
      const out = await parseAdobeBulletin(url, { vendor: 'Adobe', product: 'Adobe Commerce' });
      if (out) results.push(out);
    } catch (e) {
      console.warn(`[warn] Commerce parse failed ${url}: ${e.message}`);
    }
  }
  return results;
}

async function scrapeAEM() {
  const listUrl = 'https://helpx.adobe.com/security/products/experience-manager.html';
  const res = await fetch(listUrl, { headers: { 'user-agent': 'adobe-digest-bot/0.1 (+https://adobedigest.com)' } });
  if (!res.ok) throw new Error(`Failed to fetch ${listUrl}: ${res.status}`);
  const html = await res.text();
  const $ = cheerio.load(html);

  const links = [];
  $('a').each((_, a) => {
    const href = $(a).attr('href') || '';
    const text = $(a).text().trim();
    if (/\/security\/products\/experience-manager\//.test(href) && /APSB/i.test(text)) {
      const url = new URL(href, listUrl).toString();
      links.push({ url, text });
    }
  });

  const results = [];
  for (const { url } of links.slice(0, 30)) {
    try {
      const out = await parseAdobeBulletin(url, { vendor: 'Adobe', product: 'AEM' });
      if (out) results.push(out);
    } catch (e) {
      console.warn(`[warn] AEM parse failed ${url}: ${e.message}`);
    }
  }
  return results;
}

async function scrapeAEP() {
  // AEP (Adobe Experience Platform) bulletins are under /security/products/experience-platform.html
  const listUrl = 'https://helpx.adobe.com/security/products/experience-platform.html';
  const res = await fetch(listUrl, { headers: { 'user-agent': 'adobe-digest-bot/0.1 (+https://adobedigest.com)' } });
  if (!res.ok) throw new Error(`Failed to fetch ${listUrl}: ${res.status}`);
  const html = await res.text();
  const $ = cheerio.load(html);

  const links = [];
  $('a').each((_, a) => {
    const href = $(a).attr('href') || '';
    const text = $(a).text().trim();
    if (/\/security\/products\/experience-platform\//.test(href) && /APSB/i.test(text)) {
      const url = new URL(href, listUrl).toString();
      links.push({ url, text });
    }
  });

  const results = [];
  for (const { url } of links.slice(0, 30)) {
    try {
      const out = await parseAdobeBulletin(url, { vendor: 'Adobe', product: 'AEP' });
      if (out) results.push(out);
    } catch (e) {
      console.warn(`[warn] AEP parse failed ${url}: ${e.message}`);
    }
  }
  return results;
}

async function parseAdobeBulletin(url, { vendor, product }) {
  const res = await fetch(url, { headers: { 'user-agent': 'adobe-digest-bot/0.1 (+https://adobedigest.com)' } });
  if (!res.ok) throw new Error(`Fetch failed ${url}: ${res.status}`);
  const html = await res.text();
  const $ = cheerio.load(html);

  const title = $('h1').first().text().trim() || $('title').text().trim();
  // Try to find publication date patterns commonly in Adobe bulletins
  let dateText = $("time[datetime]").attr('datetime') || $('meta[name="publicationDate"]').attr('content') || '';
  if (!dateText) {
    const possible = $('p, span, div').filter((_, el) => /released on|published/i.test($(el).text()));
    if (possible.length) dateText = possible.first().text().replace(/.*?(\w+ \d{1,2}, \d{4}).*/, '$1');
  }
  const date = dateText && !isNaN(Date.parse(dateText)) ? new Date(dateText).toISOString() : new Date().toISOString();

  // Extract CVEs present in page text
  const text = $.root().text();
  const cves = Array.from(new Set((text.match(/CVE-\d{4}-\d{4,7}/g) || [])));

  // Severity: attempt to parse from tables or text
  let severity = 'unknown';
  const sevMatch = text.match(/(Critical|Important|Moderate|Low)/i);
  if (sevMatch) severity = sevMatch[1];

  // Body: include a short summary and link back
  const body = [
    `Source: ${url}`,
    '',
    'This page is an automated capture of an Adobe security bulletin. Refer to the source for authoritative details.',
  ].join('\n');

  return writeMarkdown({ vendor, product, title, date, severity, cves, url, body });
}

async function main() {
  const tasks = [scrapeAdobeCommerce(), scrapeAEM(), scrapeAEP()];
  const settled = await Promise.allSettled(tasks);
  let wrote = 0, skipped = 0;
  for (const s of settled) {
    if (s.status === 'fulfilled') {
      for (const r of s.value) {
        if (!r) continue;
        if (r.skipped) skipped++; else wrote++;
      }
    } else {
      console.warn('[warn] task failed:', s.reason?.message || s.reason);
    }
  }
  console.log(`Done. wrote=${wrote} skipped=${skipped}`);
}

main().catch(e => {
  console.error(e);
  process.exit(1);
});
