#!/usr/bin/env python3
"""
PDF Post-Processor - Remove waypoint name overlays

Removes BT..ET text blocks that are waypoint overlays (~44pt font size).
The 327.7pt blocks are kept - they're used for internal rendering.

Usage:
    python pdf_fixup_threshold.py <input.pdf> [--dry-run]

Examples:
    python pdf_fixup_threshold.py chart.pdf --dry-run   # Preview
    python pdf_fixup_threshold.py chart.pdf             # Apply
"""

import re
import sys
import zlib
import shutil
import argparse
from pathlib import Path

MIN_SIZE = 30.0
MAX_SIZE = 100.0

CID_TO_CHAR = {
    0x0003: " ",
    0x0024: "A",
    0x0025: "B",
    0x0026: "C",
    0x0027: "D",
    0x0028: "E",
    0x0029: "F",
    0x002A: "G",
    0x002B: "H",
    0x002C: "I",
    0x002D: "J",
    0x002E: "K",
    0x002F: "L",
    0x0030: "M",
    0x0031: "N",
    0x0032: "O",
    0x0033: "P",
    0x0035: "R",
    0x0036: "S",
    0x0037: "T",
    0x0038: "U",
    0x0039: "V",
    0x003A: "W",
    0x003B: "X",
    0x003C: "Y",
    0x003D: "Z",
    0x0013: "0",
    0x0014: "1",
    0x0015: "2",
    0x0016: "3",
    0x0017: "4",
    0x0018: "5",
    0x0019: "6",
    0x001A: "7",
    0x001B: "8",
    0x001C: "9",
    0x001D: ":",
    0x000F: ",",
    0x0010: "-",
    0x0011: ".",
    0x0012: "/",
    0x000D: "*",
    0x0008: "%",
    0x000A: "'",
    0x000B: "(",
    0x000C: ")",
    0x0041: "^",
    0x003E: "[",
    0x0040: "]",
}


def decode_text(block_content):
    result = []
    for m in re.finditer(r"<([0-9A-Fa-f]+)>", block_content):
        for i in range(0, len(m.group(1)), 4):
            if i + 4 <= len(m.group(1)):
                cid = int(m.group(1)[i : i + 4], 16)
                result.append(CID_TO_CHAR.get(cid, f"[{cid:04X}]"))
    return "".join(result)


def find_streams(data):
    streams = []
    obj_pattern = re.compile(rb"(\d+)\s+(\d+)\s+obj\s*(.*?)endobj", re.DOTALL)

    for m in obj_pattern.finditer(data):
        obj_num = int(m.group(1))
        obj_data = m.group(3)

        if b"stream" not in obj_data:
            continue

        length_m = re.search(rb"/Length\s+(\d+)", obj_data)
        if not length_m:
            continue

        declared_length = int(length_m.group(1))

        stream_start = data.find(b"stream\n", m.start())
        if stream_start == -1:
            stream_start = data.find(b"stream\r\n", m.start())
            if stream_start == -1:
                continue
            stream_start += 8
        else:
            stream_start += 7

        stream_end = data.find(b"endstream", stream_start)
        if stream_end == -1:
            continue

        actual_length = stream_end - stream_start
        raw = data[stream_start : stream_start + min(declared_length, actual_length)]
        is_compressed = b"FlateDecode" in obj_data

        streams.append(
            {
                "obj_num": obj_num,
                "obj_start": m.start(),
                "obj_end": m.end(),
                "stream_start": stream_start,
                "stream_end": stream_end,
                "raw": raw,
                "compressed": is_compressed,
                "obj_match": m,
            }
        )

    return streams


def decompress(raw, compressed):
    if compressed:
        try:
            return zlib.decompress(raw)
        except:
            pass
    return raw


def analyze_blocks(content, page_bounds):
    text = content.decode("latin-1")
    blocks = []

    for m in re.finditer(r"BT(.*?)ET", text, re.DOTALL):
        block = m.group(1)

        tf_m = re.search(r"/F\d+\s+([\d.]+)\s+Tf", block)
        if not tf_m:
            continue

        size = float(tf_m.group(1))
        decoded = decode_text(block)

        tm_m = re.search(r"1\s+0\s+0(?:\.0+)?\s+-1\s+([\d.-]+)\s+([\d.-]+)\s+Tm", block)

        tm_x, tm_y = None, None
        on_page = True

        if tm_m:
            tm_x, tm_y = float(tm_m.group(1)), float(tm_m.group(2))
            on_page = (
                page_bounds[0] - 50 <= tm_x <= page_bounds[2] + 50
                and page_bounds[1] - 50 <= tm_y <= page_bounds[3] + 50
            )

        is_overlay = MIN_SIZE <= size <= MAX_SIZE and on_page

        blocks.append(
            {
                "start": m.start(),
                "end": m.end(),
                "size": size,
                "decoded": decoded,
                "tm_x": tm_x,
                "tm_y": tm_y,
                "on_page": on_page,
                "is_overlay": is_overlay,
            }
        )

    return blocks


def analyze_image_bounds(data, page_bounds):
    """
    Finds the actual content bounds in the PDF drawing commands.
    Looks for drawing commands (m, l, re, etc) and calculates the bounding box.
    """
    streams = find_streams(data)
    min_x, min_y = float("inf"), float("inf")
    max_x, max_y = float("-inf"), float("-inf")

    # regex for coordinates in drawing commands, very simplified
    # looks for coordinates followed by m, l, re, v, y, c
    coord_pattern = re.compile(rb"(-?\d+(?:\.\d+)?)\s+(-?\d+(?:\.\d+)?)\s+(?:m|l)\b")
    rect_pattern = re.compile(
        rb"(-?\d+(?:\.\d+)?)\s+(-?\d+(?:\.\d+)?)\s+(-?\d+(?:\.\d+)?)\s+(-?\d+(?:\.\d+)?)\s+re\b"
    )

    found_content = False

    for stream in streams:
        content = decompress(stream["raw"], stream["compressed"])
        if not content:
            continue

        for match in coord_pattern.finditer(content):
            x, y = float(match.group(1)), float(match.group(2))
            min_x = min(min_x, x)
            min_y = min(min_y, y)
            max_x = max(max_x, x)
            max_y = max(max_y, y)
            found_content = True

        for match in rect_pattern.finditer(content):
            x, y, w, h = [float(match.group(i)) for i in range(1, 5)]
            min_x = min(min_x, x, x + w)
            min_y = min(min_y, y, y + h)
            max_x = max(max_x, x, x + w)
            max_y = max(max_y, y, y + h)
            found_content = True

    if not found_content:
        return None

    # Add a small margin
    margin = 5
    return [
        max(0, min_x - margin),
        max(0, min_y - margin),
        max_x + margin,
        max_y + margin,
    ]


def process_pdf(pdf_path, dry_run=False):
    pdf_path = Path(pdf_path)
    if not pdf_path.exists():
        print(f"Error: File not found: {pdf_path}")
        return 1

    with open(pdf_path, "rb") as f:
        data = f.read()

    mb_m = re.search(
        rb"/MediaBox\s*\[\s*([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s+([\d.-]+)\s*\]", data
    )
    page_bounds = [0, 0, 612, 792]
    if mb_m:
        page_bounds = [float(mb_m.group(i)) for i in range(1, 5)]

    print(f"\n{'=' * 70}")
    print(f"PDF: {pdf_path.name}")
    print(
        f"Original Page: {page_bounds[0]:.0f},{page_bounds[1]:.0f} to {page_bounds[2]:.0f},{page_bounds[3]:.0f}"
    )

    print(f"Overlay detection: {MIN_SIZE}pt <= size <= {MAX_SIZE}pt AND on-page")
    print(f"{'=' * 70}\n")

    streams = find_streams(data)
    all_blocks = []
    blocks_by_stream = {}

    for stream in streams:
        content = decompress(stream["raw"], stream["compressed"])
        if not content or b"BT" not in content:
            continue

        blocks = analyze_blocks(content, page_bounds)
        all_blocks.extend(blocks)

        overlays = [b for b in blocks if b["is_overlay"]]
        if overlays:
            blocks_by_stream[stream["obj_num"]] = {
                "stream": stream,
                "content": content,
                "blocks": overlays,
            }

    sizes = {}
    for b in all_blocks:
        s = round(b["size"], 1)
        if s not in sizes:
            sizes[s] = {"count": 0, "samples": [], "overlay": False}
        sizes[s]["count"] += 1
        if len(sizes[s]["samples"]) < 3:
            sizes[s]["samples"].append(b["decoded"][:30])
        if b["is_overlay"]:
            sizes[s]["overlay"] = True

    print("FONT SIZE DISTRIBUTION:")
    print("-" * 70)
    for size in sorted(sizes.keys(), reverse=True):
        info = sizes[size]
        if info["overlay"]:
            marker = " <-- WAYPOINT OVERLAY (REMOVE)"
        elif size > MAX_SIZE:
            marker = " [internal, kept]"
        else:
            marker = ""
        samples = "', '".join(info["samples"])
        print(f"  {size:6.1f}pt: {info['count']:3d} blocks  ['{samples}']{marker}")

    blocks_to_remove = [b for info in blocks_by_stream.values() for b in info["blocks"]]

    if blocks_to_remove:
        print(f"\n{'=' * 70}")
        print(f"BLOCKS TO REMOVE ({len(blocks_to_remove)} total):")
        print(f"{'=' * 70}")

        for b in blocks_to_remove:
            pos = (
                f"Tm=({b['tm_x']:.0f},{b['tm_y']:.0f})"
                if b["tm_x"] is not None
                else "TD"
            )
            print(f'  {b["size"]:.1f}pt {pos}: "{b["decoded"]}"')

    print(f"\n{'=' * 70}")
    print(f"SUMMARY:")
    print(f"  Total blocks:    {len(all_blocks)}")
    print(f"  To remove:       {len(blocks_to_remove)}")
    print(f"  To keep:         {len(all_blocks) - len(blocks_to_remove)}")

    if dry_run:
        print(f"\n  [DRY RUN - no changes made]")
        return 0

    if not blocks_to_remove:
        print(f"\n  No blocks to remove.")
        return 0

    print(f"\n  Applying block removals...")

    new_objects = {}
    for obj_num, info in blocks_by_stream.items():
        stream = info["stream"]
        content = info["content"]
        text = content.decode("latin-1")

        for b in sorted(info["blocks"], key=lambda x: x["start"], reverse=True):
            start, end = b["start"], b["end"]
            while start > 0 and text[start - 1] in " \t\r\n":
                start -= 1
            text = text[:start] + text[end:]

        new_content = text.encode("latin-1")
        new_raw = zlib.compress(new_content)
        new_objects[obj_num] = {
            "stream": stream,
            "new_raw": new_raw,
            "new_length": len(new_raw),
        }

    result = bytearray()
    pos = 0

    for stream in sorted(streams, key=lambda s: s["obj_start"]):
        if stream["obj_num"] in new_objects:
            result.extend(data[pos : stream["obj_start"]])

            new_info = new_objects[stream["obj_num"]]
            new_raw = new_info["new_raw"]
            new_length = new_info["new_length"]

            obj_decl = data[stream["obj_start"] : stream["stream_start"] - 7]
            length_m = re.search(rb"/Length\s+\d+", obj_decl)
            if length_m:
                new_obj_decl = (
                    obj_decl[: length_m.start()]
                    + f"/Length {new_length}".encode()
                    + obj_decl[length_m.end() :]
                )
            else:
                new_obj_decl = obj_decl

            result.extend(new_obj_decl)
            result.extend(b"\nstream\n")
            result.extend(new_raw)
            result.extend(b"\nendstream\nendobj")

            pos = stream["obj_end"]
        else:
            pass

    result.extend(data[pos:])

    backup_path = pdf_path.with_suffix(".bak")
    shutil.copy2(pdf_path, backup_path)
    print(f"  Backup: {backup_path}")

    with open(pdf_path, "wb") as f:
        f.write(result)

    print(f"  Output: {pdf_path}")
    print(f"  Removed: {len(blocks_to_remove)} blocks")

    return 0


def main():
    parser = argparse.ArgumentParser(description="Remove waypoint overlays from PDF")
    parser.add_argument("input", help="Input PDF file")
    parser.add_argument(
        "--dry-run", action="store_true", help="Preview without modifying"
    )

    args = parser.parse_args()
    return process_pdf(args.input, dry_run=args.dry_run)


if __name__ == "__main__":
    sys.exit(main())
