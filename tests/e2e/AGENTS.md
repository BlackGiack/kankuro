# Model-driven E2E runs

Before running an E2E test that contacts a model provider, ask the user which
provider/model to use and confirm a total token budget for the intended runs,
including retries. Explain estimated input and output usage before starting.
Honor an existing explicit approval covering that scope; a saved test default
or an earlier run's model choice alone is not approval for future runs.

For the full native journey, the initial planning estimate is 20,000–100,000
input tokens and 1,000–5,000 output tokens per successful run. This is an
unmeasured estimate, not a guaranteed limit: tool calls and resume requests
resend context, and retries add usage. Prefer actual usage data when available,
identify cached/reasoning tokens when reported, and label unknown usage honestly.
Do not assume a free provider means zero token consumption.

Ask again before changing an unapproved provider/model or exceeding the agreed
run/retry scope or token budget. Report observed usage when available and say
when only an estimate is available.

Ordinary tests, compile-only checks and TestNativeCaptureGlyphs make no model
requests and do not require this confirmation. Keep model-driven tests behind
the e2e build tag. Follow the root instruction not to commit unless asked.
