# Auto Deploy Detection Engine E2E Test

Use `test-auto-deploy-detection-engine.ps1` to exercise the auto-deploy API and the detection engine image end to end.

## Prerequisites
- Docker Desktop running and sharing the repository path.
- Authentication service on `http://localhost:8082` and provisioning service on `http://localhost:8083` (docker-compose defaults).
- Admin credentials `admin/password123`.
- PowerShell 7+.

## What the script does
- Ensures `iaas-detection-engine:e2e` image exists (builds from `../detection-engine` if missing).
- Uses the `/api/v1/deploy` endpoint to auto-provision ClickHouse/PostgreSQL and start the detection engine with rules mounted from `../rules_storage/detection_engine_rules.yaml`.
- Waits for the engine health endpoint, loads rules, inserts a test log, triggers detection, and fetches alerts.
- Verifies ClickHouse connectivity via `/api/v1/clickhouse/{id}/query`.

## Run
```powershell
cd vcs-infrastructure-provisioning-service
pwsh .\test-auto-deploy-detection-engine.ps1
```

## Cleanup
The script leaves the test containers running so you can inspect them. Remove them when finished (detection engine container plus the ClickHouse/PostgreSQL resources created by the deploy call).
