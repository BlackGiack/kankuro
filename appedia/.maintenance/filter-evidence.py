#!/usr/bin/env python3
"""Omit audited test bodies from published evidence, preserving run provenance."""
import argparse
import json
from pathlib import Path
import re

parser = argparse.ArgumentParser(description=__doc__)
parser.add_argument('--evidence', required=True, type=Path)
parser.add_argument('--report', type=Path, default=Path('tests_to_remove.md'))
parser.add_argument('--content', type=Path, default=Path('appedia/data/content.js'))
parser.add_argument('--out', required=True, type=Path)
args = parser.parse_args()
data = json.loads(args.evidence.read_text())
# Never edit recorded source or rehash it as if it were the original file.
excluded = set(re.findall(r'\]\(([^)]+_test\.go)#L\d+\)', args.report.read_text()))
if not excluded:
    parser.error('no test finding links found; review the report format')
removed = 0
metadata_only = 0
content = json.loads(args.content.read_text().split('=', 1)[1].strip().removesuffix(';'))
retained_files = {item['file'] for item in content.get('tests', {}).get('evidence', [])}
finding_names = set(re.findall(r'^### \d+\. (Test\w+)', args.report.read_text(), re.MULTILINE))
if finding_names.intersection(item['name'] for item in content.get('tests', {}).get('evidence', [])):
    parser.error('a reported test is still used as curated evidence')
for endpoint in ('head', 'base'):
    files = data.get(endpoint, {}).get('files', {})
    for path in excluded:
        record = files.pop(path, None)
        if record is not None:
            removed += 1
            if path in retained_files:
                files[path] = {'sha256': record['sha256'],
                               'sourceOmittedReason': 'Contains an audited test body; useful neighboring tests retain original hash and local line references only.'}
                metadata_only += 1
# Fingerprints describe the source endpoint actually extracted/executed, not
# the subset of files published. Never re-stamp coverage from another run.
args.out.parent.mkdir(parents=True, exist_ok=True)
args.out.write_text(json.dumps(data, ensure_ascii=False, indent=2) + '\n')
print(f'omitted {removed} full test-source files ({metadata_only} retain hash metadata only); endpoint and run provenance preserved')
