import argparse
import base64
import csv
import html
import json
import random
import unicodedata
from pathlib import Path

from playwright.sync_api import sync_playwright

ROOT = Path(__file__).resolve().parent.parent

LABELS = {
    "en": {
        "summary": "Summary",
        "skills": "Skills",
        "experience": "Experience",
        "education": "Education",
        "languages": "Languages",
        "certifications": "Certifications",
        "period": "Period",
        "position": "Position",
        "details": "Responsibilities",
        "contact": "Contact",
    },
    "pt": {
        "summary": "Resumo",
        "skills": "Competências",
        "experience": "Experiência Profissional",
        "education": "Formação Académica",
        "languages": "Idiomas",
        "certifications": "Certificações",
        "period": "Período",
        "position": "Função",
        "details": "Responsabilidades",
        "contact": "Contacto",
    },
}

SCANNED_ARCHETYPES = ["platform_devops", "platform_devops", "backend_go"]

LAYOUT_COUNTS = {"single": 100, "two_column": 25, "table": 22}

STYLE = """
@page { size: A4; margin: 0; }
* { box-sizing: border-box; }
body { margin: 0; padding: 16mm 18mm; font-family: Arial, Helvetica, sans-serif; font-size: 10pt; line-height: 1.4; color: #222; }
h1 { margin: 0; font-size: 22pt; }
h2 { margin: 16px 0 6px; padding-bottom: 3px; border-bottom: 1px solid #999; font-size: 11pt; text-transform: uppercase; letter-spacing: 0.05em; }
.headline { margin: 2px 0; font-size: 12pt; color: #444; }
.contact { color: #555; }
ul { margin: 4px 0 0; padding-left: 18px; }
li { margin-bottom: 2px; }
.job { margin-bottom: 10px; }
.job-head { display: flex; justify-content: space-between; }
.period { color: #555; }
table { width: 100%; border-collapse: collapse; }
td, th { padding: 5px 7px; border: 1px solid #bbb; vertical-align: top; text-align: left; }
th { background: #eee; }
.columns { display: flex; gap: 8mm; }
.sidebar { width: 62mm; flex: none; padding: 4mm; background: #eef1f5; }
.sidebar p { overflow-wrap: anywhere; }
.sidebar h2 { margin-top: 12px; }
.sidebar h2:first-child { margin-top: 0; }
.main { flex: 1; }
.scan-page { margin: 0; padding: 0; }
.scan-page img { display: block; width: 100%; filter: grayscale(1) contrast(1.15) blur(0.4px); transform: rotate(0.35deg); }
"""


def esc(value):
    return html.escape(str(value))


def ascii_name(name):
    folded = unicodedata.normalize("NFKD", name).encode("ascii", "ignore").decode()
    return " ".join(folded.split())


def wrap(body):
    return f"<!DOCTYPE html><html><head><meta charset='utf-8'><style>{STYLE}</style></head><body>{body}</body></html>"


def bullet_list(items):
    return "<ul>" + "".join(f"<li>{esc(item)}</li>" for item in items) + "</ul>"


def header(p):
    return (
        f"<h1>{esc(p['name'])}</h1>"
        f"<div class='headline'>{esc(p['headline'])}</div>"
        f"<div class='contact'>{esc(p['email'])} · {esc(p['location'])}</div>"
    )


def jobs(p):
    blocks = []
    for job in p["experience"]:
        blocks.append(
            "<div class='job'>"
            f"<div class='job-head'><span><strong>{esc(job['role'])}</strong>, {esc(job['company'])}</span>"
            f"<span class='period'>{esc(job['period'])}</span></div>"
            f"{bullet_list(job['bullets'])}</div>"
        )
    return "".join(blocks)


def education(p):
    return bullet_list(
        [f"{e['degree']}, {e['institution']} ({e['period']})" for e in p["education"]]
    )


def render_single(p):
    t = LABELS[p["language"]]
    body = header(p)
    body += f"<h2>{t['summary']}</h2><p>{esc(p['summary'])}</p>"
    body += f"<h2>{t['skills']}</h2><p>{esc(', '.join(p['skills']))}</p>"
    body += f"<h2>{t['experience']}</h2>{jobs(p)}"
    body += f"<h2>{t['education']}</h2>{education(p)}"
    body += f"<h2>{t['languages']}</h2><p>{esc(', '.join(p['languages']))}</p>"
    if p["certifications"]:
        body += f"<h2>{t['certifications']}</h2>{bullet_list(p['certifications'])}"
    return wrap(body)


def render_two_column(p):
    t = LABELS[p["language"]]
    sidebar = f"<h2>{t['contact']}</h2><p>{esc(p['email'])}<br>{esc(p['location'])}</p>"
    sidebar += f"<h2>{t['skills']}</h2>{bullet_list(p['skills'])}"
    sidebar += f"<h2>{t['languages']}</h2>{bullet_list(p['languages'])}"
    if p["certifications"]:
        sidebar += f"<h2>{t['certifications']}</h2>{bullet_list(p['certifications'])}"
    sidebar += f"<h2>{t['education']}</h2>{education(p)}"
    main = f"<h1>{esc(p['name'])}</h1><div class='headline'>{esc(p['headline'])}</div>"
    main += f"<h2>{t['summary']}</h2><p>{esc(p['summary'])}</p>"
    main += f"<h2>{t['experience']}</h2>{jobs(p)}"
    return wrap(
        f"<div class='columns'><div class='sidebar'>{sidebar}</div><div class='main'>{main}</div></div>"
    )


def render_table(p):
    t = LABELS[p["language"]]
    body = header(p)
    body += f"<h2>{t['summary']}</h2><p>{esc(p['summary'])}</p>"
    skill_rows = ""
    for i in range(0, len(p["skills"]), 3):
        cells = "".join(f"<td>{esc(s)}</td>" for s in p["skills"][i : i + 3])
        cells += "<td></td>" * (3 - len(p["skills"][i : i + 3]))
        skill_rows += f"<tr>{cells}</tr>"
    body += f"<h2>{t['skills']}</h2><table>{skill_rows}</table>"
    rows = "".join(
        f"<tr><td>{esc(j['period'])}</td><td><strong>{esc(j['role'])}</strong><br>{esc(j['company'])}</td>"
        f"<td>{bullet_list(j['bullets'])}</td></tr>"
        for j in p["experience"]
    )
    body += (
        f"<h2>{t['experience']}</h2><table><tr><th>{t['period']}</th><th>{t['position']}</th>"
        f"<th>{t['details']}</th></tr>{rows}</table>"
    )
    edu_rows = "".join(
        f"<tr><td>{esc(e['period'])}</td><td>{esc(e['degree'])}</td><td>{esc(e['institution'])}</td></tr>"
        for e in p["education"]
    )
    body += f"<h2>{t['education']}</h2><table>{edu_rows}</table>"
    body += f"<h2>{t['languages']}</h2><p>{esc(', '.join(p['languages']))}</p>"
    if p["certifications"]:
        body += f"<h2>{t['certifications']}</h2>{bullet_list(p['certifications'])}"
    return wrap(body)


RENDERERS = {
    "single": render_single,
    "two_column": render_two_column,
    "table": render_table,
}


def load_profiles(directory):
    profiles = []
    for path in sorted(directory.glob("*.json")):
        profiles.extend(json.loads(path.read_text(encoding="utf-8")))
    return profiles


def pick_scanned(profiles):
    picked = []
    for archetype in SCANNED_ARCHETYPES:
        for p in profiles:
            if p["archetype"] == archetype and not p["misleading"] and p not in picked:
                picked.append(p)
                break
    return picked


def write_pdf(page, content, path):
    page.set_content(content)
    page.pdf(path=str(path), format="A4", print_background=True, prefer_css_page_size=True)


def write_scanned_pdf(page, p, path):
    page.set_content(render_single(p))
    png = page.screenshot(full_page=True, type="png")
    encoded = base64.b64encode(png).decode()
    page.set_content(
        wrap(f"<div class='scan-page'><img src='data:image/png;base64,{encoded}'></div>")
    )
    page.pdf(path=str(path), format="A4", print_background=True, prefer_css_page_size=True)


def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("--profiles", type=Path, default=ROOT / "profiles")
    parser.add_argument("--output", type=Path, default=ROOT / "pdfs")
    parser.add_argument("--scanned-output", type=Path, default=ROOT / "pdfs-scanned")
    parser.add_argument("--manifest", type=Path, default=ROOT / "manifest.csv")
    parser.add_argument("--seed", type=int, default=7)
    args = parser.parse_args()

    profiles = load_profiles(args.profiles)
    rng = random.Random(args.seed)
    rng.shuffle(profiles)

    scanned = pick_scanned(profiles)
    regular = [p for p in profiles if p not in scanned]

    layouts = [name for name, count in LAYOUT_COUNTS.items() for _ in range(count)]
    if len(layouts) != len(regular):
        raise SystemExit(f"layout count {len(layouts)} does not match {len(regular)} profiles")
    rng.shuffle(layouts)

    args.output.mkdir(parents=True, exist_ok=True)
    args.scanned_output.mkdir(parents=True, exist_ok=True)

    rows = []
    with sync_playwright() as playwright:
        browser = playwright.chromium.launch()
        page = browser.new_page(viewport={"width": 794, "height": 1123}, device_scale_factor=2)

        for p, layout in zip(regular, layouts):
            file_name = f"{ascii_name(p['name'])}.pdf"
            write_pdf(page, RENDERERS[layout](p), args.output / file_name)
            rows.append((file_name, p, layout, "pdfs"))

        for p in scanned:
            file_name = f"{ascii_name(p['name'])}.pdf"
            write_scanned_pdf(page, p, args.scanned_output / file_name)
            rows.append((file_name, p, "scanned", "pdfs-scanned"))

        browser.close()

    rows.sort(key=lambda row: (row[3], row[0]))
    with args.manifest.open("w", newline="", encoding="utf-8") as handle:
        writer = csv.writer(handle)
        writer.writerow(["folder", "file", "archetype", "language", "layout", "misleading"])
        for file_name, p, layout, folder in rows:
            writer.writerow([folder, file_name, p["archetype"], p["language"], layout, p["misleading"]])

    print(f"{len(regular)} pdfs in {args.output}")
    print(f"{len(scanned)} scanned pdfs in {args.scanned_output}")


if __name__ == "__main__":
    main()
