#!/usr/bin/env python3
"""
Enhanced .frm table structure recovery tool for MySQL/MariaDB.
Detects likely column names and data types using heuristics.
Works on Python 3, no external dependencies.
"""

import struct
import sys
import os
import re

def extract_table_name(path):
    """Extracts the table name from the file name."""
    return os.path.splitext(os.path.basename(path))[0]

def read_frm(path):
    """Reads binary content of .frm file."""
    with open(path, "rb") as f:
        data = f.read()
    if len(data) < 100:
        raise ValueError("File too short to be a valid .frm")
    return data

def find_ascii_sequences(data):
    """Find ASCII words that could represent identifiers."""
    ascii_text = re.findall(rb"[A-Za-z0-9_]{2,30}", data)
    return [w.decode(errors="ignore") for w in ascii_text]

def guess_column_types(words):
    """
    Guess SQL column types based on nearby patterns or names.
    Uses basic keyword heuristics.
    """
    type_keywords = {
        "id": "INT",
        "num": "INT",
        "count": "INT",
        "price": "DECIMAL(10,2)",
        "amount": "DECIMAL(10,2)",
        "date": "DATETIME",
        "time": "DATETIME",
        "flag": "TINYINT(1)",
        "status": "TINYINT(1)",
        "name": "VARCHAR(255)",
        "title": "VARCHAR(255)",
        "email": "VARCHAR(255)",
        "desc": "TEXT",
        "text": "TEXT",
        "json": "JSON",
        "data": "BLOB"
    }

    columns = []
    seen = set()
    for w in words:
        wl = w.lower()
        if wl in seen or len(wl) < 2 or len(wl) > 30:
            continue
        seen.add(wl)
        if wl in {"table", "engine", "auto_increment", "default", "primary", "key", "utf8"}:
            continue
        ctype = "VARCHAR(255)"
        for key, val in type_keywords.items():
            if key in wl:
                ctype = val
                break
        columns.append((w, ctype))
    return columns[:30]

def reconstruct_create(table_name, columns):
    """Reconstruct a pseudo CREATE TABLE statement."""
    lines = [f"CREATE TABLE `{table_name}` ("]
    for col, ctype in columns:
        lines.append(f"  `{col}` {ctype} DEFAULT NULL,")
    if len(lines) > 1:
        lines[-1] = lines[-1].rstrip(',')
    lines.append(") ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;")
    return "\n".join(lines)

def main(path):
    print(f"# Reading: {path}")
    data = read_frm(path)
    words = find_ascii_sequences(data)
    columns = guess_column_types(words)
    table_name = extract_table_name(path)
    sql = reconstruct_create(table_name, columns)
    print(sql)

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python3 frm_reader_plus.py /path/to/table.frm")
        sys.exit(1)
    main(sys.argv[1])
