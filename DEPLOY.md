# Deploy to Production

Complete guide to deploy your Pack Calculator as a live, publicly accessible application.

---

## Quick Deploy Options

### 🚀 Fastest: Railway (5 minutes)

**Best for:** Quick deployment, automatic HTTPS, free tier available

1. Push code to GitHub:
   ```bash
   git init
   git add .
   git commit -m "Initial commit"
   git remote add origin YOUR_GITHUB_URL
   git push -u origin main
   ```

2. Deploy to Railway:
   - Go to [railway.app](https://railway.app)
   - Click "Start a New Project"
   - Select "Deploy from GitHub repo"
   - Choose your repository
   - Railway auto-detects `docker-compose.yml`
   - Click "Deploy"

3. Add PostgreSQL:
   - Click "New" → "Database" → "Add PostgreSQL"
   - Railway automatically connects it

4. Get your URL:
   - Settings → Generate Domain
   - Your app: `https://your-app.railway.app`

**Cost:** Free tier (500 hours/month, $5 credit)

---

### ⚡ Easy: Render (10 minutes)

**Best for:** Simple setup, free tier, automatic SSL

1. Push to GitHub (same as above)

2. Deploy Backend:
   - Go to [render.com](https://render.com)
   - Click "New +" → "Web Service"
   - Connect GitHub repo
   - Settings:
     - Name: `pack-calculator-backend`
     - Environment: `Docker`
     - Docker Command: (auto-detected)
     - Instance Type: `Free`
   - Click "Create Web Service"

3. Add PostgreSQL:
   - Click "New +" → "PostgreSQL"
   - Name: `pack-calculator-db`
   - Plan: `Free`
   - Click "Create Database"

4. Connect Database:
   - Go to backend service
   - Environment tab
   - Add from `pack-calculator-db`:
     - `DATABASE_URL` (internal connection)

5. Deploy Frontend:
   - Click "New +" → "Static Site"
   - Connect same repo
   - Settings:
     - Build Command: `cd frontend && npm install && npm run build`
     - Publish Directory: `frontend/build`
   - Environment:
     - `REACT_APP_API_URL`: Your backend URL
   - Click "Create Static Site"

**Cost:** Free tier (750 hours/month)

---

### 🐳 Advanced: DigitalOcean App Platform (15 minutes)

**Best for:** More control, scalable, $5/month

1. Push to GitHub

2. Create App:
   - Go to [digitalocean.com/products/app-platform](https://www.digitalocean.com/products/app-platform)
   - Click "Create App"
   - Connect GitHub repo
   - Choose branch: `main`

3. Configure Services:
   - **Backend:**
     - Type: Web Service
     - Dockerfile: `backend/Dockerfile`
     - HTTP Port: 8080
     - Instance Size: Basic ($5/month)
   
   - **Frontend:**
     - Type: Static Site
     - Build Command: `cd frontend && npm install && npm run build`
     - Output Directory: `frontend/build`
   
   - **Database:**
     - Type: Dev Database (PostgreSQL)
     - Or: Managed Database ($15/month for production)

4. Set Environment Variables:
   - Backend:
     ```
     DB_HOST=${db.HOSTNAME}
     DB_PORT=${db.PORT}
     DB_USER=${db.USERNAME}
     DB_PASSWORD=${db.PASSWORD}
     DB_NAME=${db.DATABASE}
     PORT=8080
     CACHE_SIZE=1000
     ```
   
   - Frontend:
     ```
     REACT_APP_API_URL=${backend.PUBLIC_URL}
     ```

5. Deploy:
   - Click "Create Resources"
   - Wait 5-10 minutes
   - Your app: `https://your-app.ondigitalocean.app`

**Cost:** $5-20/month depending on resources

---

### 🌐 Full Control: AWS ECS/Fargate (30 minutes)

**Best for:** Enterprise, full control, scalable

See detailed AWS guide below.

---

## Detailed Deployment Guides

### Option 1: Railway (Detailed)

#### Prerequisites
- GitHub account
- Railway account (free)

#### Steps

**1. Prepare Repository**
```bash
# Initialize git
git init

# Create .gitignore (if not exists)
cat > .gitignore << 'EOF'
node_modules/
*.log
.env
.DS_Store
EOF

# Commit code
git add .
git commit -m "Deploy to Railway"

# Push to GitHub
gh repo create pack-calculator --public --source=. --remote=origin --push
# Or manually:
# git remote add origin https://github.com/YOUR_USERNAME/pack-calculator.git
# git push -u origin main
```

**2. Deploy on Railway**

1. **Sign in:** https://railway.app → Login with GitHub

2. **New Project:**
   - Click "New Project"
   - Select "Deploy from GitHub repo"
   - Authorize Railway to access your repos
   - Choose `pack-calculator`

3. **Configure Services:**
   Railway auto-detects `docker-compose.yml` and creates:
   - ✅ Backend service
   - ✅ Frontend service
   - ✅ PostgreSQL database

4. **Set Variables** (if needed):
   - Click backend service → Variables
   - Add:
     ```
     CACHE_SIZE=1000
     API_KEY=your-secret-key-123 (optional)
     ```

5. **Generate Domain:**
   - Click backend → Settings → Generate Domain
   - Get URL: `https://pack-calculator-backend-production-xxxx.up.railway.app`
   - Click frontend → Settings → Generate Domain
   - Get URL: `https://pack-calculator-frontend-production-xxxx.up.railway.app`

6. **Update Frontend Env:**
   - Click frontend → Variables
   - Add:
     ```
     REACT_APP_API_URL=https://your-backend-url.up.railway.app
     ```
   - Redeploy frontend

**3. Verify Deployment**
```bash
# Test backend
curl https://your-backend-url.up.railway.app/health

# Open frontend
open https://your-frontend-url.up.railway.app
```

**4. Custom Domain (Optional)**
- Settings → Domains → Add Custom Domain
- Add CNAME record in your DNS:
  ```
  pack-calculator CNAME your-app.up.railway.app
  ```

**Cost:** 
- Free: $5 credit, 500 hours/month
- Pro: $20/month, $0.000463/GB-hour

---

### Option 2: Render (Detailed)

#### Prerequisites
- GitHub account
- Render account (free)

#### Steps

**1. Push to GitHub** (same as Railway)

**2. Create PostgreSQL Database**

1. Go to https://render.com → Dashboard
2. Click "New +" → "PostgreSQL"
3. Settings:
   - Name: `pack-calculator-db`
   - Database: `packcalculator`
   - User: `pack_user`
   - Region: Choose closest to you
   - Plan: `Free`
4. Click "Create Database"
5. Save these from Info tab:
   - Internal Database URL
   - External Database URL

**3. Deploy Backend**

1. Click "New +" → "Web Service"
2. Connect repository: `pack-calculator`
3. Settings:
   - Name: `pack-calculator-backend`
   - Region: Same as database
   - Branch: `main`
   - Root Directory: `backend`
   - Environment: `Docker`
   - Docker Build Context Directory: `backend`
   - Docker Command: (auto-detected from Dockerfile)
   - Instance Type: `Free`

4. Environment Variables (click "Add Environment Variable"):
   ```
   PORT=8080
   DB_HOST=dpg-xxxxx-a.oregon-postgres.render.com
   DB_PORT=5432
   DB_USER=pack_user
   DB_PASSWORD=xxxxx (from database info)
   DB_NAME=packcalculator
   CACHE_SIZE=1000
   ```
   
   Or use Internal Database URL:
   ```
   DATABASE_URL=postgresql://pack_user:password@dpg-xxxxx:5432/packcalculator
   ```
   
   Then modify `main.go` to parse `DATABASE_URL` if provided.

5. Click "Create Web Service"
6. Wait 5-10 minutes for build
7. Your backend: `https://pack-calculator-backend.onrender.com`

**4. Deploy Frontend**

1. Click "New +" → "Static Site"
2. Connect same repository
3. Settings:
   - Name: `pack-calculator-frontend`
   - Branch: `main`
   - Root Directory: `frontend`
   - Build Command: `npm install && npm run build`
   - Publish Directory: `build`

4. Environment Variables:
   ```
   REACT_APP_API_URL=https://pack-calculator-backend.onrender.com
   ```

5. Click "Create Static Site"
6. Wait 5 minutes
7. Your frontend: `https://pack-calculator-frontend.onrender.com`

**5. Test Application**
```bash
# Backend health
curl https://pack-calculator-backend.onrender.com/health

# Backend calculation
curl -X POST https://pack-calculator-backend.onrender.com/api/calculate \
  -H "Content-Type: application/json" \
  -d '{"amount": 501}'

# Frontend
open https://pack-calculator-frontend.onrender.com
```

**Important Notes:**
- Free tier spins down after 15 minutes of inactivity
- First request after spin-down takes ~30 seconds
- Upgrade to paid tier ($7/month) for always-on

**Cost:**
- Free: Limited hours
- Starter: $7/month per service
- Pro: $25/month per service

---

### Option 3: DigitalOcean App Platform (Detailed)

#### Prerequisites
- DigitalOcean account
- GitHub account
- Credit card (required even for free tier)

#### Steps

**1. Create App**

1. Go to https://cloud.digitalocean.com/apps
2. Click "Create App"
3. Choose Source: GitHub
4. Authorize DigitalOcean
5. Select repository: `pack-calculator`
6. Select branch: `main`
7. Click "Next"

**2. Configure Resources**

DigitalOcean auto-detects `docker-compose.yml`. Edit each:

**Backend Service:**
```yaml
Name: backend
Type: Web Service
Source Directory: /backend
Dockerfile Path: backend/Dockerfile
HTTP Port: 8080
HTTP Request Routes: /
Health Check: /health

Environment Variables:
  PORT=8080
  DB_HOST=${db.HOSTNAME}
  DB_PORT=${db.PORT}
  DB_USER=${db.USERNAME}
  DB_PASSWORD=${db.PASSWORD}
  DB_NAME=${db.DATABASE}
  CACHE_SIZE=1000

Resources:
  Instance Type: Basic
  Instance Size: $5/month (512MB RAM, 1 vCPU)
```

**Frontend Service:**
```yaml
Name: frontend
Type: Static Site
Source Directory: /frontend
Build Command: npm install && npm run build
Output Directory: build

Environment Variables:
  REACT_APP_API_URL=${backend.PUBLIC_URL}

Resources:
  Static site (Free)
```

**Database:**
```yaml
Name: db
Type: Database
Engine: PostgreSQL
Version: 15
Plan: Dev Database (Free) or Basic ($15/month)
```

**3. Deploy**

1. Click "Next" → "Next"
2. Review settings
3. Click "Create Resources"
4. Wait 10-15 minutes for deployment
5. Your app: `https://pack-calculator-xxxx.ondigitalocean.app`

**4. Custom Domain**

1. Go to Settings → Domains
2. Click "Add Domain"
3. Enter: `pack-calculator.yourdomain.com`
4. Add DNS records (shown in UI):
   ```
   Type: CNAME
   Host: pack-calculator
   Value: pack-calculator-xxxx.ondigitalocean.app
   ```
5. SSL certificate auto-generated

**5. Monitor**

- Insights tab: CPU, Memory, Requests
- Runtime Logs: Real-time logs
- Deployments: History and rollback

**Cost:**
- Dev Database: Free (25MB storage)
- Basic App: $5/month (512MB RAM)
- Pro App: $12/month (1GB RAM)
- Database (Managed): $15/month (1GB RAM)

---

### Option 4: Heroku (Simple but Paid)

#### Prerequisites
- Heroku account
- Heroku CLI

#### Steps

```bash
# Install Heroku CLI
brew tap heroku/brew && brew install heroku

# Login
heroku login

# Create app
heroku create pack-calculator-backend

# Add PostgreSQL
heroku addons:create heroku-postgresql:mini

# Set environment variables
heroku config:set CACHE_SIZE=1000
heroku config:set PORT=8080

# Deploy backend
git subtree push --prefix backend heroku main

# Or use Docker
heroku container:push web --app pack-calculator-backend
heroku container:release web --app pack-calculator-backend

# View logs
heroku logs --tail

# Open app
heroku open
```

**Frontend on Heroku:**
```bash
# Create frontend app
heroku create pack-calculator-frontend --buildpack mars/create-react-app

# Set backend URL
heroku config:set REACT_APP_API_URL=https://pack-calculator-backend.herokuapp.com

# Deploy
git subtree push --prefix frontend heroku main
```

**Cost:**
- Eco Dynos: $5/month per dyno
- Mini Postgres: $5/month
- Total: ~$15/month

---

### Option 5: AWS ECS with Fargate (Advanced)

#### Prerequisites
- AWS account
- AWS CLI configured
- Docker installed

#### Steps

**1. Setup AWS Resources**

```bash
# Install AWS CLI
brew install awscli

# Configure
aws configure

# Create ECR repositories
aws ecr create-repository --repository-name pack-calculator-backend
aws ecr create-repository --repository-name pack-calculator-frontend

# Get ECR login
aws ecr get-login-password --region us-east-1 | \
  docker login --username AWS --password-stdin \
  ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com
```

**2. Build and Push Images**

```bash
# Backend
cd backend
docker build -t pack-calculator-backend .
docker tag pack-calculator-backend:latest \
  ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/pack-calculator-backend:latest
docker push ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/pack-calculator-backend:latest

# Frontend
cd ../frontend
docker build -t pack-calculator-frontend .
docker tag pack-calculator-frontend:latest \
  ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/pack-calculator-frontend:latest
docker push ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/pack-calculator-frontend:latest
```

**3. Create RDS PostgreSQL**

```bash
# Via AWS Console or CLI
aws rds create-db-instance \
  --db-instance-identifier pack-calculator-db \
  --db-instance-class db.t3.micro \
  --engine postgres \
  --master-username postgres \
  --master-user-password YOUR_PASSWORD \
  --allocated-storage 20
```

**4. Create ECS Cluster**

1. Go to AWS ECS Console
2. Create Cluster:
   - Name: `pack-calculator-cluster`
   - Infrastructure: AWS Fargate
3. Create Task Definition (backend):
   - Family: `pack-calculator-backend`
   - Container:
     - Image: `ACCOUNT_ID.dkr.ecr.us-east-1.amazonaws.com/pack-calculator-backend:latest`
     - Port: 8080
     - Environment:
       - DB_HOST: RDS endpoint
       - DB_PORT: 5432
       - DB_USER: postgres
       - DB_PASSWORD: YOUR_PASSWORD
       - DB_NAME: packcalculator
   - Resources:
     - CPU: 0.5 vCPU
     - Memory: 1GB

4. Create Service:
   - Cluster: `pack-calculator-cluster`
   - Launch type: Fargate
   - Task: `pack-calculator-backend`
   - Desired tasks: 2
   - Load Balancer: Application Load Balancer
   - Target Group: Create new

5. Repeat for frontend

**Cost:**
- Fargate: ~$15-30/month (0.5 vCPU, 1GB)
- RDS: ~$15-25/month (db.t3.micro)
- ALB: ~$20/month
- Total: ~$50-75/month

---

### Option 6: Google Cloud Run (Serverless)

#### Steps

```bash
# Install gcloud CLI
brew install --cask google-cloud-sdk

# Initialize
gcloud init

# Enable services
gcloud services enable run.googleapis.com
gcloud services enable sql-component.googleapis.com

# Create PostgreSQL instance
gcloud sql instances create pack-calculator-db \
  --database-version=POSTGRES_15 \
  --tier=db-f1-micro \
  --region=us-central1

# Create database
gcloud sql databases create packcalculator --instance=pack-calculator-db

# Deploy backend
cd backend
gcloud run deploy pack-calculator-backend \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated \
  --set-env-vars="DB_HOST=/cloudsql/PROJECT_ID:REGION:pack-calculator-db"

# Deploy frontend
cd ../frontend
gcloud run deploy pack-calculator-frontend \
  --source . \
  --platform managed \
  --region us-central1 \
  --allow-unauthenticated

# Get URLs
gcloud run services list
```

**Cost:**
- Cloud Run: Pay per request (free tier: 2M requests/month)
- Cloud SQL: ~$10/month (db-f1-micro)
- Total: ~$10-20/month

---

## Environment Variables Summary

### Backend
```bash
PORT=8080
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your-password
DB_NAME=packcalculator
CACHE_SIZE=1000
API_KEY=your-secret-key-optional
```

### Frontend
```bash
REACT_APP_API_URL=https://your-backend-url.com
```

---

## Post-Deployment Checklist

### ✅ Verification

1. **Backend Health:**
   ```bash
   curl https://your-backend-url.com/health
   ```
   Expected: `{"status":"healthy","cache":{...}}`

2. **Backend Calculation:**
   ```bash
   curl -X POST https://your-backend-url.com/api/calculate \
     -H "Content-Type: application/json" \
     -d '{"amount": 501}'
   ```
   Expected: `{"amount":501,"total_items":750,"total_packs":2,"packs":{"250":1,"500":1}}`

3. **Frontend:**
   - Open in browser
   - Test calculation: 501 → 750 items (1×500 + 1×250)
   - Check pack management
   - Check order history

4. **Edge Case:**
   - Change pack sizes to: 23, 31, 53
   - Calculate: 500,000
   - Expected: {23: 2, 31: 7, 53: 9429}

### ✅ Performance

```bash
# Load test
for i in {1..100}; do
  curl -s -o /dev/null -w "%{time_total}\n" \
    -X POST https://your-backend-url.com/api/calculate \
    -H "Content-Type: application/json" \
    -d '{"amount": 1000}' &
done | awk '{sum+=$1; count++} END {print "Average:", sum/count, "seconds"}'
```

Expected: <0.1 seconds average

### ✅ Monitoring

1. **Check Logs:**
   - Railway: Deployments tab → Logs
   - Render: Logs tab
   - DigitalOcean: Runtime Logs
   - AWS: CloudWatch

2. **Monitor Metrics:**
   - Response times
   - Error rates
   - Memory usage
   - CPU usage
   - Database connections

3. **Setup Alerts:**
   - Uptime monitoring (UptimeRobot, Pingdom)
   - Error tracking (Sentry)
   - Performance monitoring (New Relic, Datadog)

---

## Custom Domain Setup

### Namecheap/GoDaddy/Any Registrar

1. **Buy domain:** `pack-calculator.com`

2. **Add DNS records:**
   ```
   # For Railway/Render/DO
   Type: CNAME
   Host: @
   Value: your-app-url.platform.com
   TTL: Auto
   
   # For AWS with ALB
   Type: A (Alias)
   Host: @
   Value: your-alb-endpoint.elb.amazonaws.com
   ```

3. **Configure platform:**
   - Railway: Settings → Custom Domain → Add
   - Render: Settings → Custom Domain → Add
   - DigitalOcean: Settings → Domains → Add

4. **Wait for SSL:**
   - Auto-generated Let's Encrypt certificate
   - Usually takes 5-15 minutes

---

## Cost Comparison

| Platform | Free Tier | Paid (Basic) | Paid (Production) |
|----------|-----------|--------------|-------------------|
| **Railway** | $5 credit, 500 hrs | $20/month | $50+/month |
| **Render** | Limited hours | $7/service | $25/service |
| **DigitalOcean** | Static site only | $5/month | $20-50/month |
| **Heroku** | None | $5/dyno | $25+/month |
| **AWS ECS** | 12 months free | $30/month | $75+/month |
| **Google Cloud Run** | 2M requests free | $10/month | $30+/month |

---

## Recommendations

### For Learning/Demo: **Railway** or **Render**
- ✅ Free tier
- ✅ 5-minute setup
- ✅ Automatic HTTPS
- ✅ Easy to use

### For Production: **DigitalOcean** or **AWS**
- ✅ More control
- ✅ Better performance
- ✅ Scalable
- ✅ Professional features

### For Serverless: **Google Cloud Run**
- ✅ Pay per use
- ✅ Auto-scaling
- ✅ Cost-effective for low traffic

---

## Troubleshooting

### Database Connection Errors

```bash
# Check database is accessible
pg_isready -h your-db-host -p 5432

# Test connection
psql postgres://user:pass@host:5432/dbname

# Check environment variables
echo $DB_HOST
```

### Frontend Can't Reach Backend

```bash
# Check CORS headers
curl -I https://your-backend-url.com/api/calculate

# Should include:
# Access-Control-Allow-Origin: *
```

### Slow First Request (Render Free Tier)

- Free tier spins down after 15 minutes
- First request takes 30-60 seconds
- Solution: Upgrade to paid tier or use cron job to ping every 10 minutes

### Build Failures

```bash
# Check build logs
# Ensure all dependencies in package.json/go.mod
# Verify Dockerfile syntax
# Test build locally first:
docker build -t test ./backend
```

---

## Next Steps After Deployment

1. **Setup Monitoring:**
   - UptimeRobot for uptime
   - Sentry for errors
   - Google Analytics for usage

2. **Performance:**
   - Enable CDN (Cloudflare)
   - Add Redis for caching
   - Optimize images

3. **Security:**
   - Enable rate limiting
   - Add API key auth
   - Setup WAF

4. **CI/CD:**
   - GitHub Actions auto-deploy
   - Automated tests
   - Staging environment

---

**Your app is now live! 🚀**

Share your URL: `https://your-pack-calculator.com`

