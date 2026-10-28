# SC common-use Han list

`gb2312-han.txt` contains exactly 6,763 unique Unicode scalars, one hexadecimal
codepoint per line, sorted. Source: the Han rows (16–87, encoded byte pairs
B0A1–F7FE) of GB 2312-1980, decoded with Python's standard-library `gb2312`
codec. Unassigned positions are skipped. No frequency corpus is redistributed.

The encoding is documented in the [Python codec table](https://docs.python.org/3/library/codecs.html#standard-encodings).
To reconstruct the list body independently:

```python
characters = set()
for lead in range(0xB0, 0xF8):
    for trail in range(0xA1, 0xFF):
        try:
            characters.add(ord(bytes([lead, trail]).decode("gb2312")))
        except UnicodeDecodeError:
            pass
assert len(characters) == 6763
print("".join(f"{cp:04X}\n" for cp in sorted(characters)), end="")
```

The subset script reads this checked-in list; it does not regenerate it from a
host codec. U+6F22 漢 is a separately documented extra in the recipe, not a
member of this list.
