import requests
from bs4 import BeautifulSoup

URL = "https://helpx.adobe.com/security/products/magento.html"
HEADERS = {
    "User-Agent": "Mozilla/5.0 (compatible; adobe-digest-bot/0.1)",
    "Accept": "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8",
    "Accept-Language": "en-US,en;q=0.9",
    "Upgrade-Insecure-Requests": "1",
}

def main():
    resp = requests.get(URL, headers=HEADERS, timeout=20)
    resp.raise_for_status()
    soup = BeautifulSoup(resp.text, "html.parser")
    links = []
    for a in soup.find_all("a", href=True):
        href = a["href"]
        text = a.get_text(strip=True)
        if "/security/products/magento/" in href and "APSB" in text:
            full_url = requests.compat.urljoin(URL, href)
            links.append(full_url)
    print(f"Found {len(links)} bulletins:")
    for url in links:
        print(url)

if __name__ == "__main__":
    main()
