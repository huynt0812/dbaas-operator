#!/bin/bash
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 DBaaS Operator Quick Start${NC}"
echo ""

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"
if ! command -v kubectl &> /dev/null; then
    echo -e "${RED}❌ kubectl not found. Please install kubectl first.${NC}"
    exit 1
fi

if ! kubectl cluster-info &> /dev/null; then
    echo -e "${RED}❌ Cannot connect to Kubernetes cluster. Check your kubeconfig.${NC}"
    exit 1
fi

echo -e "${GREEN}✅ Prerequisites OK${NC}"
echo ""

# Install CNPG operator
echo -e "${YELLOW}📦 Installing CloudNativePG operator...${NC}"
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.23/releases/cnpg-1.23.0.yaml

echo -e "${YELLOW}⏳ Waiting for CNPG operator to be ready...${NC}"
kubectl wait --for=condition=Available --timeout=300s \
  deployment/cnpg-controller-manager -n cnpg-system 2>/dev/null || {
  echo -e "${YELLOW}⚠️  CNPG operator taking longer than expected, continuing anyway...${NC}"
}

echo -e "${GREEN}✅ CNPG operator installed${NC}"
echo ""

# Generate and install CRDs
echo -e "${YELLOW}🔧 Generating CRDs...${NC}"
if command -v controller-gen &> /dev/null; then
    make manifests
    echo -e "${GREEN}✅ CRDs generated${NC}"
else
    echo -e "${YELLOW}⚠️  controller-gen not found, using existing CRDs${NC}"
fi

echo -e "${YELLOW}📋 Installing DBaaS CRDs...${NC}"
# Check if CRD directory exists, if not create minimal CRD files
if [ ! -d "config/crd/bases" ]; then
    echo -e "${YELLOW}⚠️  CRD directory not found, skipping CRD installation${NC}"
    echo -e "${YELLOW}   Run 'make manifests' to generate CRDs${NC}"
else
    kubectl apply -f config/crd/bases/
    echo -e "${GREEN}✅ CRDs installed${NC}"
fi
echo ""

# Create namespace
echo -e "${YELLOW}🏗️  Creating dbaas-system namespace...${NC}"
kubectl create namespace dbaas-system --dry-run=client -o yaml | kubectl apply -f -
echo -e "${GREEN}✅ Namespace created${NC}"
echo ""

# Apply RBAC
echo -e "${YELLOW}🔐 Applying RBAC...${NC}"
kubectl apply -f config/rbac/role.yaml
echo -e "${GREEN}✅ RBAC configured${NC}"
echo ""

# Deploy operator
echo -e "${YELLOW}🚢 Deploying DBaaS operator...${NC}"
kubectl apply -f config/manager/manager.yaml

echo -e "${YELLOW}⏳ Waiting for operator to be ready...${NC}"
kubectl wait --for=condition=Available --timeout=300s \
  deployment/dbaas-operator-controller-manager -n dbaas-system 2>/dev/null || {
  echo -e "${YELLOW}⚠️  Operator taking longer than expected, check status manually${NC}"
}

echo -e "${GREEN}✅ Operator deployed${NC}"
echo ""

# Create test resources
echo -e "${YELLOW}📝 Creating test resources...${NC}"
kubectl apply -f config/samples/databaseengine-cnpg.yaml
echo -e "${GREEN}✅ DatabaseEngine created${NC}"

kubectl apply -f config/samples/postgresql-cluster.yaml
echo -e "${GREEN}✅ PostgreSQL cluster created${NC}"
echo ""

# Show status
echo -e "${GREEN}🎉 Installation complete!${NC}"
echo ""
echo -e "${YELLOW}Check cluster status with:${NC}"
echo "  kubectl get databasecluster postgresql-demo -w"
echo ""
echo -e "${YELLOW}View operator logs:${NC}"
echo "  kubectl logs -n dbaas-system deployment/dbaas-operator-controller-manager -f"
echo ""
echo -e "${YELLOW}View created resources:${NC}"
echo "  kubectl get databasecluster"
echo "  kubectl get cluster"
echo "  kubectl get pods -l dbaas.io/cluster=postgresql-demo"
echo ""
echo -e "${YELLOW}Try some operations:${NC}"
echo "  kubectl apply -f config/samples/ops-horizontal-scaling.yaml"
echo "  kubectl apply -f config/samples/ops-backup.yaml"
echo ""
