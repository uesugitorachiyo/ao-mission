# AO Stack Public Release Stabilization Handoff

This preserves the executable handoff and terminal closure of the public-stack
stabilization goal. Do not restart Windows remediation or rewrite its retained
evidence.

```text
Audit the completed Codex goal:

Complete the AO Stack Windows compatibility release cycle so ao-architecture
references one coherent, publicly downloadable release set that installs and
passes clean-host verification on Ubuntu, macOS, and Windows, while preserving
repository release, signing, evidence, and authority contracts.

Terminal state on 2026-08-22:
- Windows remediation Mission: mission-4864066542d3c329, done and reconciled.
- Windows rollup SHA-256:
  96e3397f00155ded2b7e567ac668df889f5d5e2181c5e82c7ffad98166eb5629.
- AO Mission v0.1.5 is public at source
  5d4562578a4751d56910ef108b930fbb8dc91e7d; rehearsal 32532558410 and
  finalizer 32532729277 passed.
- AO Forge v0.1.5 source d1723769949269dcd0589916d83769dcb7275f98
  passed rehearsal 32532526229 and has policy-valid signed annotated tag
  v0.1.5. Promotion-plan SHA-256:
  f614c81743bf9f954bdf823bcbdfc49f5c1cc7543f48f0e8c3b5dab79157217f.
  Draft publisher 32535826667 and finalizer 32539072103 passed; the signed
  release is public.
- AO Command v0.1.3 is public at source
  ffef6d76306e892c3e7a7f39734433d5a832006a; rehearsal 32536508516 and
  plan-bound publisher 32536659576 passed, including fresh public downloads.
- AO Atlas v0.2.1 terminal source:
  3603a2bb8af5adafcd9ff17b807ab89f32283d18. Native run 32537521755,
  rehearsal 32537634684, approved manifest digest
  sha256:6b35e072aa7a1c4d07fcf546fb1af1706870a625c13e61cd838ad905ffbe881b,
  and promotion-plan digest
  03392a4ffdb197a4cfc69ac44bf0523deeac01dea449d51102b0abcb41f17bc3
  are bound. Finalizer 32537720561 published the release; PR 771 repaired the
  portable post-public verifier without replacing assets. All 15 public assets
  were independently rehashed.
- AO Architecture terminal source:
  2a0f0291166f3d98374f3d7d718ea99c975065ab.
- Merged-main public canary 32540433860 passed Linux x86_64, macOS arm64, and
  Windows x86_64 from public download URLs.
- The integrated source, native-artifact, and hosted Ubuntu/macOS/Windows
  matrices passed before merge.

Current public truth is coherent. AO Architecture references Mission v0.1.5,
Command v0.1.3, Atlas v0.2.1, and Forge v0.1.5 in
stack/current-release-manifest.json, docs/current-release.md,
docs/stack-assembly.md, and scripts/run_public_stack_canary.py.

Closure audit:
1. Preserve the four public releases and their exact tag targets unchanged.
2. Verify every public tag target, archive name, size, and SHA-256 by fresh
   download. Do not reuse workflow workspace bytes as public verification.
3. Require Architecture release truth and the public canary asset table to
   match the exact published versions and digests.
4. Require Architecture CI and Public Stack Canary 32540433860 to remain green
   on linux-x86_64, macos-aarch64, and windows-x86_64.

The set is stable for the documented supported targets. Preserve completed
historical evidence. Keep provider execution, deployment, promotion,
compatibility activation, and RSI authority denied. Release authority does not
grant those authorities.
```
