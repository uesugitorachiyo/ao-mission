# AO Mission v0.1 SDD

AO Mission creates durable mission records, route decisions, continuation steps, scheduler readbacks, gateway intents, artifact refs, and governance snapshots. It does not schedule as the mission brain and does not execute mutation.

## Pipeline

User, CLI, Telegram, or A2A creates an intent. AO Mission records the mission and routes underspecified work to AO Blueprint, Atlas-required work to AO Atlas, ready nodes to AO Foundry, and bounded execution packets to AO Forge/AO2 only through downstream gates.

## Completion

Handoff generation is not completion. A mission is done only when a final rollup is recorded or it is stopped with an exact blocker.
