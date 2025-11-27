# Docker-in-Docker (DinD) Service - Use Case & Architecture

## 1. Tổng quan

Docker-in-Docker cho phép user có một **môi trường Docker riêng biệt và isolated** để:
- Chạy bất kỳ docker command nào
- Build images
- Run containers
- Thực hiện docker-compose
- Testing CI/CD pipelines

## 2. Use Cases

### UC1: Tạo DinD Environment
```
User → API: POST /api/v1/dind/environments
           { "name": "my-sandbox", "resource_plan": "medium" }
           
System:
  1. Tạo docker:dind container với privileged mode
  2. Khởi động Docker daemon bên trong
  3. Trả về environment_id và connection info
```

### UC2: Chạy Docker Command
```
User → API: POST /api/v1/dind/environments/{id}/exec
           { "command": "docker run -d nginx" }
           
System:
  1. Tìm DinD container của environment
  2. Exec command vào DinD container
  3. Trả về output và exit code
```

### UC3: Build Docker Image
```
User → API: POST /api/v1/dind/environments/{id}/build
           { 
             "dockerfile": "FROM node:18\nRUN npm install",
             "image_name": "my-app",
             "tag": "latest"
           }
           
System:
  1. Copy Dockerfile vào DinD container
  2. Chạy docker build
  3. Trả về build logs và image info
```

### UC4: Run Docker Compose
```
User → API: POST /api/v1/dind/environments/{id}/compose
           {
             "compose_content": "version: '3'\nservices:\n  web:\n    image: nginx",
             "action": "up"
           }
           
System:
  1. Copy docker-compose.yml vào DinD container
  2. Chạy docker-compose up/down
  3. Trả về output
```

## 3. Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                    DinD Architecture                             │
└─────────────────────────────────────────────────────────────────┘

┌──────────────┐     ┌─────────────────────────────────────────────┐
│   User API   │     │           DinD Container                    │
│   Request    │────▶│  ┌─────────────────────────────────────┐   │
│              │     │  │     docker:dind (privileged)        │   │
│ docker run   │     │  │  ┌──────────────────────────────┐   │   │
│ docker build │     │  │  │   Isolated Docker Daemon     │   │   │
│ docker-compose│    │  │  │  ┌────┐ ┌────┐ ┌────┐       │   │   │
└──────────────┘     │  │  │  │ C1 │ │ C2 │ │ C3 │       │   │   │
                     │  │  │  └────┘ └────┘ └────┘       │   │   │
                     │  │  │   User's containers          │   │   │
                     │  │  └──────────────────────────────┘   │   │
                     │  └─────────────────────────────────────┘   │
                     └─────────────────────────────────────────────┘
                                         │
                                         ▼
                              ┌─────────────────┐
                              │  Host Docker    │
                              │  Engine         │
                              └─────────────────┘
```

## 4. Security Considerations

### Privileged Mode
- DinD **yêu cầu** `--privileged` flag
- Mỗi user có **container DinD riêng** → isolated
- Giới hạn resources (CPU, Memory)
- Network isolation

### Resource Limits
```yaml
resources:
  small:   { cpu: "1", memory: "1GB" }
  medium:  { cpu: "2", memory: "2GB" }  
  large:   { cpu: "4", memory: "4GB" }
```

## 5. API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /dind/environments | Tạo DinD environment mới |
| GET | /dind/environments | List environments của user |
| GET | /dind/environments/{id} | Lấy thông tin environment |
| DELETE | /dind/environments/{id} | Xóa environment |
| POST | /dind/environments/{id}/exec | Chạy docker command |
| POST | /dind/environments/{id}/build | Build Docker image |
| POST | /dind/environments/{id}/compose | Run docker-compose |
| GET | /dind/environments/{id}/images | List images trong environment |
| GET | /dind/environments/{id}/containers | List containers trong environment |
| GET | /dind/environments/{id}/logs | Lấy logs của environment |

## 6. Data Flow

```
1. User tạo environment
   POST /dind/environments
        ↓
2. System tạo docker:dind container
   docker run -d --privileged --name dind-{id} docker:dind
        ↓
3. User chạy command
   POST /dind/environments/{id}/exec { "command": "docker run nginx" }
        ↓
4. System exec vào DinD container
   docker exec dind-{id} docker run nginx
        ↓
5. Trả về output cho user
```

## 7. Example Usage

### Tạo environment và chạy nginx
```bash
# 1. Tạo environment
curl -X POST http://localhost:8083/api/v1/dind/environments \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name": "my-sandbox", "resource_plan": "medium"}'

# Response: {"environment_id": "abc-123", ...}

# 2. Chạy nginx trong environment
curl -X POST http://localhost:8083/api/v1/dind/environments/abc-123/exec \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"command": "docker run -d nginx"}'

# 3. List containers
curl -X GET http://localhost:8083/api/v1/dind/environments/abc-123/containers \
  -H "Authorization: Bearer $TOKEN"
```

### Build và run custom image
```bash
# 1. Build image
curl -X POST http://localhost:8083/api/v1/dind/environments/abc-123/build \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "dockerfile": "FROM node:18\nWORKDIR /app\nCOPY . .\nRUN npm install\nCMD [\"npm\", \"start\"]",
    "image_name": "my-node-app",
    "tag": "v1"
  }'

# 2. Run built image
curl -X POST http://localhost:8083/api/v1/dind/environments/abc-123/exec \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"command": "docker run -d my-node-app:v1"}'
```

