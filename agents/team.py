"""Agent team definitions for the Prestia multi-agent system."""

from claude_agent_sdk import AgentDefinition

from prompts import AGENT_PROMPTS

AGENT_TOOLS = {
    "architect": ["Read", "Glob", "Grep", "Write", "Edit"],
    "backend": ["Read", "Glob", "Grep", "Write", "Edit", "Bash"],
    "frontend": ["Read", "Glob", "Grep", "Write", "Edit", "Bash"],
    "database": ["Read", "Glob", "Grep", "Write", "Edit", "Bash"],
    "qa": ["Read", "Glob", "Grep", "Write", "Edit", "Bash"],
    "devops": ["Read", "Glob", "Grep", "Write", "Edit", "Bash"],
}

AGENT_DESCRIPTIONS = {
    "architect": (
        "System architect for hexagonal architecture design, "
        "domain modeling, and interface definitions"
    ),
    "backend": (
        "Go backend developer for Gin/GORM microservices, "
        "REST API endpoints, and JWT authentication"
    ),
    "frontend": (
        "React frontend developer for Material Design UI, "
        "Astro static pages, and client portal"
    ),
    "database": (
        "Database specialist for PostgreSQL schema design "
        "via GORM AutoMigrate, migrations, and seed data"
    ),
    "qa": (
        "QA engineer for unit tests on domain models, "
        "integration tests for services, and test fixtures"
    ),
    "devops": (
        "DevOps engineer for Docker, docker-compose, "
        "GCP deployment configs, and CI/CD pipelines"
    ),
}


def build_team() -> dict[str, AgentDefinition]:
    """Build and return the dictionary of all 6 subagent definitions."""
    agents = {}
    for name in AGENT_PROMPTS:
        agents[name] = AgentDefinition(
            description=AGENT_DESCRIPTIONS[name],
            prompt=AGENT_PROMPTS[name],
            tools=AGENT_TOOLS[name],
        )
    return agents
