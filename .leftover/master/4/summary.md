# Summary (2026-03-11)
Goal: improve session switching performance and workspace correctness for Kankuro MVP.

Status: implemented output handling fixes (reduce cross-session bleed, cap no-newline buffer), and ensured working directory is honored via PWD env override. User reports much better behavior, with a small remaining output buffer when switching.
