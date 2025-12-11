# Deployment Documentation

## Overview

Streamz can be deployed in several ways depending on your needs:
1. **Local Development** - Docker Compose
2. **Single VPS** - DigitalOcean, Hetzner, etc.
3. **Cloud Platform** - AWS, GCP, or Railway/Render

---

## Local Development

### Prerequisites
- Docker & Docker Compose
- Go 1.21+ (for local development without Docker)
- Node.js 18+ (for frontend development)
- PostgreSQL 14+ (or use Docker)

### Docker Compose Setup

```yaml
# docker-compose.yml
version: '3.8'

services:
  db:
    image: postgres:14-alpine
    environment:
      POSTGRES_USER: streamz
      POSTGRES_PASSWORD: streamz
      POSTGRES_DB: streamz
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data

  backend:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "8080:8080"
    environment:
      DATABASE_URL: postgres://streamz:streamz@db:5432/streamz?sslmode=disable
      JWT_SECRET: dev-secret-change-in-prod
      PORT: 8080
      ENV: development
    depends_on:
      - db

  coturn:
    image: coturn/coturn:latest
    ports:
      - "3478:3478/udp"
      - "3478:3478/tcp"
    command: >
      -n
      --log-file=stdout
      --lt-cred-mech
      --fingerprint
      --realm=streamz.local
      --user=streamz:streamz

volumes:
  postgres_data:
```

### Commands

```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f backend

# Stop services
docker-compose down

# Reset database
docker-compose down -v && docker-compose up -d
```

---

## Single VPS Deployment

### Recommended Specs
- **CPU:** 2 vCPUs
- **RAM:** 4GB
- **Storage:** 40GB SSD
- **OS:** Ubuntu 22.04 LTS

### Setup Steps

#### 1. Install Dependencies
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Install Docker
curl -fsSL https://get.docker.com | sh
sudo usermod -aG docker $USER

# Install Docker Compose
sudo apt install docker-compose-plugin

# Install Nginx (reverse proxy)
sudo apt install nginx certbot python3-certbot-nginx
```

#### 2. Configure Nginx
```nginx
# /etc/nginx/sites-available/streamz
server {
    listen 80;
    server_name streamz.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

```bash
# Enable site
sudo ln -s /etc/nginx/sites-available/streamz /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx

# Get SSL certificate
sudo certbot --nginx -d streamz.yourdomain.com
```

#### 3. Deploy Application
```bash
# Clone repository
git clone https://github.com/yourusername/streamz.git
cd streamz

# Create production env file
cp .env.example .env.prod
# Edit .env.prod with production values

# Build and run
docker-compose -f docker-compose.prod.yml up -d
```

---

## Cloud Platform Deployment

### Option 1: Railway

Railway offers simple deployment with automatic SSL and database provisioning.

```bash
# Install Railway CLI
npm install -g @railway/cli

# Login and deploy
railway login
railway init
railway add --database postgres
railway up
```

**Environment Variables (set in Railway dashboard):**
```
DATABASE_URL=${{Postgres.DATABASE_URL}}
JWT_SECRET=your-production-secret
PORT=8080
ENV=production
```

### Option 2: Render

**Backend Service:**
1. Connect GitHub repository
2. Select "Web Service"
3. Build Command: `go build -o main ./cmd/app`
4. Start Command: `./main`

**Database:**
1. Create PostgreSQL instance
2. Copy connection string to environment variables

### Option 3: AWS (ECS + RDS)

#### Infrastructure Overview
```
┌─────────────────────────────────────────────────────────┐
│                        AWS VPC                          │
├─────────────────────────────────────────────────────────┤
│  ┌─────────────┐    ┌─────────────┐    ┌─────────────┐ │
│  │   ALB       │────│   ECS       │────│   RDS       │ │
│  │  (HTTPS)    │    │  (Fargate)  │    │ (Postgres)  │ │
│  └─────────────┘    └─────────────┘    └─────────────┘ │
│                            │                            │
│                     ┌──────┴──────┐                     │
│                     │ ElastiCache │                     │
│                     │  (optional) │                     │
│                     └─────────────┘                     │
└─────────────────────────────────────────────────────────┘
```

#### ECS Task Definition
```json
{
  "family": "streamz",
  "networkMode": "awsvpc",
  "requiresCompatibilities": ["FARGATE"],
  "cpu": "512",
  "memory": "1024",
  "containerDefinitions": [
    {
      "name": "streamz",
      "image": "your-ecr-repo/streamz:latest",
      "portMappings": [
        {
          "containerPort": 8080,
          "protocol": "tcp"
        }
      ],
      "environment": [
        {"name": "ENV", "value": "production"},
        {"name": "PORT", "value": "8080"}
      ],
      "secrets": [
        {"name": "DATABASE_URL", "valueFrom": "arn:aws:ssm:..."},
        {"name": "JWT_SECRET", "valueFrom": "arn:aws:ssm:..."}
      ],
      "logConfiguration": {
        "logDriver": "awslogs",
        "options": {
          "awslogs-group": "/ecs/streamz",
          "awslogs-region": "us-east-1",
          "awslogs-stream-prefix": "ecs"
        }
      }
    }
  ]
}
```

---

## TURN Server Deployment

For WebRTC P2P fallback, you need a TURN server. Options:

### Option 1: Self-hosted Coturn
```bash
# Install on Ubuntu
sudo apt install coturn

# Configure
sudo nano /etc/turnserver.conf
```

```ini
# /etc/turnserver.conf
listening-port=3478
tls-listening-port=5349
fingerprint
lt-cred-mech
realm=streamz.yourdomain.com
user=streamz:secure-password
total-quota=100
stale-nonce=600
cert=/etc/letsencrypt/live/turn.yourdomain.com/fullchain.pem
pkey=/etc/letsencrypt/live/turn.yourdomain.com/privkey.pem
```

### Option 2: Twilio TURN
- Sign up at twilio.com
- Use their Network Traversal Service
- Get credentials via API

### Option 3: Metered TURN
- Sign up at metered.ca
- Simple API-based credentials
- Pay per usage

---

## CI/CD with GitHub Actions

```yaml
# .github/workflows/deploy.yml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  build-and-deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Set up Go
        uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run tests
        run: go test ./...

      - name: Build Docker image
        run: docker build -t streamz:${{ github.sha }} .

      - name: Push to Registry
        run: |
          echo ${{ secrets.DOCKER_PASSWORD }} | docker login -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
          docker tag streamz:${{ github.sha }} yourusername/streamz:latest
          docker push yourusername/streamz:latest

      # Add deployment steps for your platform
```

---

## Production Checklist

### Security
- [ ] Change all default passwords
- [ ] Set strong JWT_SECRET (32+ characters)
- [ ] Enable HTTPS everywhere
- [ ] Configure CORS for production domains only
- [ ] Enable rate limiting
- [ ] Set up firewall rules (ufw/security groups)

### Database
- [ ] Enable SSL connections
- [ ] Set up automated backups
- [ ] Configure connection pooling
- [ ] Create read replicas (if needed)

### Monitoring
- [ ] Set up structured logging
- [ ] Configure error tracking (Sentry)
- [ ] Set up uptime monitoring
- [ ] Configure alerts for critical errors

### Performance
- [ ] Enable gzip compression
- [ ] Configure CDN for static assets
- [ ] Set appropriate cache headers
- [ ] Tune database connection pool

---

## Environment Variables Reference

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `PORT` | No | 8080 | HTTP server port |
| `ENV` | No | development | Environment (development/production) |
| `DATABASE_URL` | Yes | - | PostgreSQL connection string |
| `JWT_SECRET` | Yes | - | Secret key for JWT signing |
| `JWT_EXPIRY` | No | 24h | JWT token expiration |
| `ALLOWED_ORIGINS` | No | * | CORS allowed origins |
| `TURN_URL` | No | - | TURN server URL |
| `TURN_USERNAME` | No | - | TURN server username |
| `TURN_PASSWORD` | No | - | TURN server password |
