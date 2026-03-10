# Decisions
- Decision: Use "word silence" (15s without alphanumeric characters) for idle detection.
  Rationale: Many interactive CLIs use animations or non-standard prompts that don't match simple character heuristics. Tracking when the last "meaningful" content was received is more robust.
  Tradeoffs: A CLI that outputs only numbers or words in an animation might still be marked as "running" longer than expected, but this is safer than marking it "waiting" while it's still outputting data.

- Decision: Remove `CaptureOutput` and the `output` pointer in `GenericCLIDriver`.
  Rationale: `Spawn` already returns the output channel, and `CaptureOutput` was unused in the codebase. Removing it simplifies the interface and avoids keeping redundant pointers.
  Tradeoffs: None, as the functionality was redundant.
