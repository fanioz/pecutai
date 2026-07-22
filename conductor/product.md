# Initial Concept
WUPHF is an open-source collaborative office for AI employees with a shared brain, designed for developers and technical founders. It brings CEO, PM, engineers, designers, and marketers into a visible shared terminal and web workspace to collaborate, claim tasks, and ship work locally.

---

# Product Guide

## Vision & Purpose
WUPHF (pronounced "woof") is a local-first, open-source collaborative office for AI employees with a shared brain running 24x7. Rather than disappearing behind a black-box API or hidden automated queue, the AI team (CEO, PM, engineers, designers, CMO, and CRO) operates in a visible, shared workspace where users can watch them argue, collaborate, divide tasks, and commit code.

The project is designed to make multi-agent software engineering tangible, debuggable, and fun. It aims to deliver a credible, robust, and local-first development system that starts with a single command: `npx wuphf`.

## User Persona
1. **Developers & Hackers**: Looking for a reliable local multi-agent setup that they can easily modify, inspect, and extend.
2. **Technical Founders**: Evaluating AI teammates for operational tasks, rapid prototyping, and automated workflows.
3. **AI Tool Builders**: Looking for a structured framework to build custom agents, integrate custom tools, and configure shared memory wikis.

## Core Features
1. **Shared Workspace & Brain**: Agents share a wiki (local git repo of markdown files at `~/.wuphf/wiki/`) and private notebooks to graduate context, observe state, and run checks.
2. **Unified Agent Team**: CEO, PM, engineers, designers, CMO, and CRO agents interact in a collaborative lobby.
3. **Local Dev Toolchain**: Built-in SQLite, Bleve text search, web mock screenshot tools, and LLM router integrations.
4. **Tailscale & Cloudflare Sharing**: Integrated tailscale-based mesh sharing (`wuphf share`) and Cloudflare quick tunnels.
5. **Skill Marketplace**: Import and export skills (`wuphf skills publish/install`) using GitHub PRs.

## Brand Voice & Personality
* **Direct & Product-First**: Speaks directly, avoid vague SaaS copy or marketing hype.
* **Playful yet Credible**: Employs self-aware humor (themed around the "office" and "WUPHF") but functions as a serious, production-grade developer tool underneath.
* **Transparent & Local**: Emphasizes git-based state, markdown files, and local execution.
