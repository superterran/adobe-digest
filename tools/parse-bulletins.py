#!/usr/bin/env python3
"""
Adobe Security Bulletin Page Parser

This script helps manually extract bulletin information from the comprehensive
Adobe security bulletin page. It generates a JSON file that can be imported
using the bulk-importer tool.

Since Adobe's CDN blocks automated scraping, this tool is designed to work with
manually copy-pasted content from https://helpx.adobe.com/security/security-bulletin.html
"""

import json
import re
import sys
from datetime import datetime
from typing import List, Dict, Any


def parse_bulletin_line(line: str) -> Dict[str, Any]:
    """
    Parse a single bulletin line like:
    "APSB25-85 - Security update available for Adobe Acrobat Reader - September 9, 2025"
    """
    # Match APSB ID, title, and date
    pattern = r'(APSB\d{2}-\d{2,3})\s*[-–]\s*(.+?)\s*[-–]\s*(.+?)$'
    match = re.match(pattern, line.strip())
    
    if not match:
        return None
    
    apsb_id = match.group(1)
    title = match.group(2).strip()
    date_str = match.group(3).strip()
    
    # Parse date (various formats possible)
    date_obj = parse_date(date_str)
    
    # Generate URL
    url = f"https://helpx.adobe.com/security/products/{infer_product_path(title)}/{apsb_id.lower()}.html"
    
    # Infer product from title
    products = infer_products(title)
    
    # Infer severity (would need more context, defaulting to Important)
    severity = "Important"
    
    return {
        "apsb": apsb_id,
        "title": f"{apsb_id}: {title}",
        "description": f"Adobe has released security updates for {', '.join(products)}. More details in the security bulletin.",
        "url": url,
        "date": date_obj.isoformat() + "Z",
        "products": products,
        "severity": severity
    }


def parse_date(date_str: str) -> datetime:
    """Parse various date formats"""
    # Try different date formats
    formats = [
        "%B %d, %Y",      # September 9, 2025
        "%b %d, %Y",      # Sep 9, 2025
        "%Y-%m-%d",       # 2025-09-09
        "%m/%d/%Y",       # 09/09/2025
    ]
    
    for fmt in formats:
        try:
            return datetime.strptime(date_str, fmt)
        except ValueError:
            continue
    
    # Default to current date if parsing fails
    return datetime.now()


def infer_product_path(title: str) -> str:
    """Infer the product path for URL generation"""
    title_lower = title.lower()
    
    if "acrobat" in title_lower:
        return "acrobat"
    elif "photoshop" in title_lower:
        return "photoshop"
    elif "after effects" in title_lower:
        return "after-effects"
    elif "illustrator" in title_lower:
        return "illustrator"
    elif "premiere" in title_lower:
        return "premiere"
    elif "lightroom" in title_lower:
        return "lightroom"
    elif "indesign" in title_lower:
        return "indesign"
    elif "dreamweaver" in title_lower:
        return "dreamweaver"
    elif "animate" in title_lower:
        return "animate"
    elif "audition" in title_lower:
        return "audition"
    elif "bridge" in title_lower:
        return "bridge"
    elif "dimension" in title_lower:
        return "dimension"
    elif "experience manager" in title_lower or "aem" in title_lower:
        return "experience-manager"
    elif "commerce" in title_lower or "magento" in title_lower:
        return "commerce"
    elif "coldfusion" in title_lower:
        return "coldfusion"
    elif "campaign" in title_lower:
        return "campaign"
    elif "substance" in title_lower:
        return "substance"
    else:
        return "other"


def infer_products(title: str) -> List[str]:
    """Infer product names from the title"""
    title_lower = title.lower()
    products = []
    
    product_mappings = {
        "acrobat": ["Adobe Acrobat", "Adobe Acrobat Reader"],
        "photoshop": ["Adobe Photoshop"],
        "after effects": ["Adobe After Effects"],
        "illustrator": ["Adobe Illustrator"],
        "premiere": ["Adobe Premiere Pro"],
        "lightroom": ["Adobe Lightroom"],
        "indesign": ["Adobe InDesign"],
        "dreamweaver": ["Adobe Dreamweaver"],
        "animate": ["Adobe Animate"],
        "audition": ["Adobe Audition"],
        "bridge": ["Adobe Bridge"],
        "dimension": ["Adobe Dimension"],
        "experience manager": ["Adobe Experience Manager"],
        "commerce": ["Adobe Commerce"],
        "magento": ["Adobe Commerce"],
        "coldfusion": ["Adobe ColdFusion"],
        "campaign": ["Adobe Campaign"],
        "substance": ["Adobe Substance 3D"],
        "creative cloud": ["Adobe Creative Cloud"]
    }
    
    for keyword, product_list in product_mappings.items():
        if keyword in title_lower:
            products.extend(product_list)
    
    # If no products found, try to extract from title
    if not products:
        products = ["Adobe Product"]
    
    return list(set(products))  # Remove duplicates


def main():
    if len(sys.argv) != 3:
        print("Adobe Security Bulletin Parser")
        print("Usage: python3 tools/parse-bulletins.py <input-file> <output-file>")
        print()
        print("Input file should contain bulletin lines like:")
        print("APSB25-85 - Security update available for Adobe Acrobat Reader - September 9, 2025")
        print("APSB25-84 - Security update available for Adobe Photoshop - September 2, 2025")
        print()
        print("You can copy these lines from https://helpx.adobe.com/security/security-bulletin.html")
        sys.exit(1)
    
    input_file = sys.argv[1]
    output_file = sys.argv[2]
    
    bulletins = []
    
    try:
        with open(input_file, 'r', encoding='utf-8') as f:
            lines = f.readlines()
        
        print(f"📄 Processing {len(lines)} lines from {input_file}")
        
        for i, line in enumerate(lines, 1):
            line = line.strip()
            if not line or not line.startswith('APSB'):
                continue
            
            bulletin = parse_bulletin_line(line)
            if bulletin:
                bulletins.append(bulletin)
                print(f"  ✅ {bulletin['apsb']}: {bulletin['title']}")
            else:
                print(f"  ⚠️  Line {i}: Could not parse: {line}")
        
        # Create output structure
        output_data = {
            "bulletins": bulletins
        }
        
        # Write JSON file
        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(output_data, f, indent=2, ensure_ascii=False)
        
        print(f"\n✅ Successfully parsed {len(bulletins)} bulletins")
        print(f"📁 Output saved to {output_file}")
        print(f"\n🔄 Next steps:")
        print(f"   1. Review the generated JSON file")
        print(f"   2. Run: go run cmd/bulk-importer/main.go data/security-bulletins.json {output_file}")
        print(f"   3. Run: go run cmd/content-generator/main.go generate")
        
    except Exception as e:
        print(f"❌ Error: {e}")
        sys.exit(1)


if __name__ == "__main__":
    main()
