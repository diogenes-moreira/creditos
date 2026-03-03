#!/usr/bin/env python3
"""CLI entry point for the Prestia multi-agent orchestrator."""

import asyncio
import sys
from pathlib import Path

from claude_agent_sdk import (
    query,
    ClaudeAgentOptions,
    CLINotFoundError,
    CLIConnectionError,
    ProcessError,
)

from prompts import ORCHESTRATOR_PROMPT
from team import build_team

PROJECT_ROOT = str(Path(__file__).resolve().parent.parent)


def build_options() -> ClaudeAgentOptions:
    """Configure the orchestrator agent options."""
    return ClaudeAgentOptions(
        cwd=PROJECT_ROOT,
        permission_mode="acceptEdits",
        allowed_tools=["Read", "Glob", "Grep", "Bash", "Task"],
        agents=build_team(),
        system_prompt=ORCHESTRATOR_PROMPT,
        model="claude-sonnet-4-6",
    )


async def run(task: str) -> None:
    """Run the orchestrator with a given task description."""
    options = build_options()

    print(f"Project root: {PROJECT_ROOT}")
    print(f"Task: {task}")
    print("=" * 72)

    async for message in query(prompt=task, options=options):
        if message.type == "assistant":
            for block in message.content:
                if hasattr(block, "text"):
                    print(block.text)
                elif hasattr(block, "name"):
                    agent = ""
                    if block.name == "Task":
                        agent = f" -> {block.input.get('subagent_type', '?')}"
                    print(f"[tool: {block.name}{agent}]")
        elif message.type == "result":
            print("=" * 72)
            print(f"Result: {message.subtype}")
            if hasattr(message, "duration_ms") and message.duration_ms:
                print(f"Duration: {message.duration_ms}ms")
            if hasattr(message, "total_cost_usd") and message.total_cost_usd:
                print(f"Cost: ${message.total_cost_usd:.4f}")
            if hasattr(message, "result") and message.result:
                print(message.result)


def main() -> None:
    if len(sys.argv) < 2:
        print("Usage: python agents/main.py \"<task description>\"")
        print()
        print("Examples:")
        print('  python agents/main.py "Design the domain model for credits"')
        print('  python agents/main.py "Set up the hexagonal architecture"')
        print('  python agents/main.py "Create the loan amortization logic"')
        sys.exit(1)

    task = " ".join(sys.argv[1:])

    try:
        asyncio.run(run(task))
    except CLINotFoundError:
        print(
            "Error: Claude Code CLI not found. "
            "Install with: pip install claude-agent-sdk",
            file=sys.stderr,
        )
        sys.exit(1)
    except CLIConnectionError as exc:
        print(f"Error: Connection failed: {exc}", file=sys.stderr)
        sys.exit(1)
    except ProcessError as exc:
        print(f"Error: Process failed: {exc}", file=sys.stderr)
        sys.exit(1)
    except KeyboardInterrupt:
        print("\nInterrupted.")
        sys.exit(130)


if __name__ == "__main__":
    main()
