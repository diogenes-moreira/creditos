# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Crédito Villanueva** — a microcredit management system for Argentina. Clients get current accounts, credit lines, and can request loans with French or German amortization. Admins approve credits, adjust payments, and manage the portfolio. Includes audit trail, PDF generation, early cancellation, and a dashboard with KPIs (delinquency, portfolio status).

## Architecture

- **Microservices** with **hexagonal architecture** (ports & adapters)
- Backend: **Go** with **Gin** (HTTP framework) and **GORM** (ORM)
- Frontend: **React** (app) + **Astro** (static content), using **Material Design**
- Database: **PostgreSQL** — use GORM's `AutoMigrate` only, no raw SQL anywhere
- Auth: **JWT** for API authentication; **Firebase Auth** for user management
- File storage: **Google Cloud Storage**
- Deployment: **Docker**, targeting GCP

## Key Conventions

- All source code and comments in **English**
- **Rich domain models** — business logic lives in the model, not anemic DTOs
- Frontend communicates exclusively through a **public REST API** with JWT
- Every domain class must have **unit tests**; every service must have **integration tests**
- All operations must be **audited**: user, timestamp, IP, user-agent, and action description
