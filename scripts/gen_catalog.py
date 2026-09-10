#!/usr/bin/env python3
"""Generate internal/stt/models.json from the handy-computer Hugging Face org.

Everything factual — size, checksum, languages, word error rate, real-time
factor — comes from the API. Only the curation below is ours. The checksum is
the trust anchor for a downloaded file, so an invented one is worse than none.
"""

import json
import os
import sys
import urllib.error
import urllib.request

API = "https://huggingface.co/api"
ORG = "handy-computer"
UA = {"User-Agent": "encre-catalog"}

QUANT_PREFERENCE = ["Q8_0", "Q5_K_M", "Q6_K", "Q4_K_M", "F16", "F32"]

# Offered first, with copy written for someone choosing rather than benchmarking.
FEATURED = {
    "parakeet-unified-en-0.6b": (1, "Fast and accurate English. The best default if you dictate in English."),
    "whisper-small": (2, "Multilingual workhorse. Good accuracy at moderate cost."),
    "whisper-tiny": (3, "Smallest multilingual model. Runs anywhere, least accurate."),
    "canary-180m-flash": (4, "English, German, Spanish and French, with translation."),
    "SenseVoiceSmall": (5, "Chinese, Cantonese, English, Japanese and Korean."),
    "whisper-medium": (6, "Higher accuracy, noticeably slower on CPU."),
}

# Too large to be a sensible download for a dictation tool.
MAX_SIZE_BYTES = 4 * 1024**3


def fetch(url):
    request = urllib.request.Request(url, headers=UA)
    with urllib.request.urlopen(request, timeout=60) as response:
        return json.load(response)


def split_camel(word):
    """SenseVoiceSmall -> Sense Voice Small, leaving acronyms and versions be."""
    out, current = [], ""
    for char in word:
        if char.isupper() and current and not current[-1].isupper():
            out.append(current)
            current = char
        else:
            current += char
    if current:
        out.append(current)
    return out


def display_name(slug):
    words = []
    for chunk in slug.replace("_", "-").split("-"):
        if any(c.isupper() for c in chunk[1:]):
            words.extend(split_camel(chunk))
        elif chunk.isupper() or any(c.isdigit() for c in chunk):
            words.append(chunk)
        else:
            words.append(chunk.capitalize())
    return " ".join(words)


def pick_quant(files):
    by_quant = {}
    for f in files:
        path = f["path"]
        if not path.endswith(".gguf"):
            continue
        for quant in QUANT_PREFERENCE:
            if path.upper().endswith(f"-{quant}.GGUF"):
                by_quant[quant] = f
    for quant in QUANT_PREFERENCE:
        if quant in by_quant:
            return quant, by_quant[quant]
    return None, None


def score_from_wer(wer):
    """WER of 0 is perfect; 30% is unusable. Clamped into 0..1."""
    if wer is None:
        return None
    return max(0.0, min(1.0, 1.0 - wer / 30.0))


def score_from_rtf(rtf):
    """Real-time factor: 40x and above reads as full marks."""
    if rtf is None:
        return None
    return max(0.0, min(1.0, rtf / 40.0))


def best_wer(meta, quant):
    for key, value in meta.items():
        if key.startswith("wer_") and isinstance(value, dict):
            return value.get(quant.lower()) or next(iter(value.values()), None)
    return None


def cpu_rtf(meta):
    """Slowest measured CPU machine. Ranking should under-promise, not over."""
    rtfs = [v.get("cpu") for k, v in meta.items() if k.startswith("rtf_") and isinstance(v, dict)]
    rtfs = [r for r in rtfs if r]
    return min(rtfs) if rtfs else None


def resolve(repo_id):
    slug = repo_id.split("/")[-1].removesuffix("-gguf")

    try:
        info = fetch(f"{API}/models/{repo_id}")
    except urllib.error.HTTPError as e:
        print(f"  skip {slug}: HTTP {e.code}", file=sys.stderr)
        return None

    revision = info.get("sha")
    card = info.get("cardData") or {}
    meta = card.get("transcribe_cpp") or {}

    try:
        files = fetch(f"{API}/models/{repo_id}/tree/{revision}?recursive=1")
    except urllib.error.HTTPError:
        return None

    quant, chosen = pick_quant(files)
    if chosen is None:
        print(f"  skip {slug}: no usable quantisation", file=sys.stderr)
        return None

    lfs = chosen.get("lfs") or {}
    sha256 = lfs.get("oid")
    size = lfs.get("size") or chosen.get("size")

    if not sha256:
        print(f"  skip {slug}: no checksum published", file=sys.stderr)
        return None
    if size and size > MAX_SIZE_BYTES:
        print(f"  skip {slug}: {size / 1024**3:.1f} GB is too large", file=sys.stderr)
        return None

    languages = card.get("language") or []
    if isinstance(languages, str):
        languages = [languages]

    wer = best_wer(meta, quant)
    rtf = cpu_rtf(meta)
    rank, description = FEATURED.get(slug, (None, None))

    if description is None:
        if len(languages) == 1:
            description = f"{display_name(slug)}, {languages[0]} only."
        elif languages:
            description = f"{display_name(slug)}, {len(languages)} languages."
        else:
            description = display_name(slug)

    return {
        "id": f"{repo_id}/{chosen['path']}",
        "slug": slug,
        "name": display_name(slug),
        "description": description,
        "repo": repo_id,
        "revision": revision,
        "filename": chosen["path"],
        "quant": quant,
        "size_bytes": size,
        "sha256": sha256,
        "languages": languages,
        "license": card.get("license", ""),
        "translate": bool(meta.get("translate")),
        "streaming": bool(meta.get("streaming")),
        "language_detect": bool(meta.get("lang_detect")),
        "word_error_rate": wer,
        "realtime_factor": rtf,
        "speed_score": score_from_rtf(rtf) or 0.5,
        "accuracy_score": score_from_wer(wer) or 0.5,
        "recommended": rank is not None,
        "rank": rank or 999,
    }


def main():
    print(f"enumerating {ORG}...")
    repos = fetch(f"{API}/models?author={ORG}&limit=500")
    gguf = sorted(r["id"] for r in repos if r["id"].endswith("-gguf"))
    print(f"{len(gguf)} gguf repositories\n")

    models = []
    for repo_id in gguf:
        row = resolve(repo_id)
        if row:
            models.append(row)
            print(f"  {row['slug']:<34} {row['size_bytes'] / 1048576:7.1f} MB  {row['quant']:<7}"
                  f" {len(row['languages']) or '?':>3} lang")

    if not models:
        sys.exit("no models resolved; refusing to write an empty catalog")

    models.sort(key=lambda m: (m["rank"], -m["accuracy_score"]))

    catalog = {
        "catalog_version": 2,
        "source": f"https://huggingface.co/{ORG}",
        "models": models,
    }

    root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    out = os.path.join(root, "internal", "stt", "models.json")
    with open(out, "w") as handle:
        json.dump(catalog, handle, indent=2)
        handle.write("\n")

    print(f"\nwrote {len(models)} models to {out}")


if __name__ == "__main__":
    main()
