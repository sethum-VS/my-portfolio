#!/usr/bin/env bash
# =============================================================================
#  azure-setup.sh — One-shot provisioning: Heroku → Azure Container Apps
#
#  Run ONCE locally after `az login` and `az account set`.
#  After this script completes you will have:
#    • Resource group + Container Apps environment
#    • Container App with native secrets (no plain-text env vars in ARM)
#    • Service Principal credentials → paste as AZURE_CREDENTIALS secret
#    • DNS instructions + domain-binding commands for sethum.dev
#
#  Prerequisites:
#    brew install azure-cli jq
#    az login
#    az extension add --name containerapp --upgrade -y
# =============================================================================
set -euo pipefail

# ─── Colour helpers ──────────────────────────────────────────────────────────
RED='\033[0;31m'; YELLOW='\033[1;33m'; GREEN='\033[0;32m'
CYAN='\033[0;36m'; BOLD='\033[1m'; RESET='\033[0m'
info()    { echo -e "${CYAN}[INFO]${RESET}  $*"; }
success() { echo -e "${GREEN}[OK]${RESET}    $*"; }
warn()    { echo -e "${YELLOW}[WARN]${RESET}  $*"; }
banner()  { echo -e "\n${BOLD}${CYAN}══════════════════════════════════════════${RESET}"; echo -e "${BOLD}${CYAN}  $*${RESET}"; echo -e "${BOLD}${CYAN}══════════════════════════════════════════${RESET}\n"; }

# ─── Configuration ────────────────────────────────────────────────────────────
# Adjust these if needed. Region is centralindia — the only unblocked region
# for this Azure for Students subscription. Conveniently also close to IST timezone.
RESOURCE_GROUP="rg-portfolio"
LOCATION="centralindia"
ENVIRONMENT="cae-portfolio"
APP_NAME="ca-portfolio"
TARGET_PORT="80"
PLACEHOLDER_IMAGE="mcr.microsoft.com/k8se/quickstart:latest"
DOMAIN="sethum.dev"
SP_NAME="sp-portfolio-github"
SUBSCRIPTION_ID="$(az account show --query id -o tsv)"

# ─── Secret placeholder values ───────────────────────────────────────────────
# Replace these with your real values before running, OR set the corresponding
# environment variables in your shell before executing the script.
SUPABASE_DB_URL="${SUPABASE_DB_URL:-REPLACE_WITH_YOUR_SUPABASE_DB_URL}"
SUPABASE_ANON_KEY="${SUPABASE_ANON_KEY:-REPLACE_WITH_YOUR_SUPABASE_ANON_KEY}"
SUPABASE_JWT_SECRET="${SUPABASE_JWT_SECRET:-REPLACE_WITH_YOUR_SUPABASE_JWT_SECRET}"
SUPABASE_SERVICE_ROLE_KEY="${SUPABASE_SERVICE_ROLE_KEY:-REPLACE_WITH_YOUR_SUPABASE_SERVICE_ROLE_KEY}"
SUPABASE_URL="${SUPABASE_URL:-REPLACE_WITH_YOUR_SUPABASE_URL}"
RESEND_API_KEY="${RESEND_API_KEY:-REPLACE_WITH_YOUR_RESEND_API_KEY}"
TURNSTILE_SECRET_KEY="${TURNSTILE_SECRET_KEY:-REPLACE_WITH_YOUR_TURNSTILE_SECRET_KEY}"
NVIDIA_NIM_API="${NVIDIA_NIM_API:-REPLACE_WITH_YOUR_NVIDIA_NIM_API}"
# GCP_CREDENTIALS_FILE: path to your GCP service account JSON file.
# The script reads this file, single-lines it with jq, and stores it as an
# Azure secret — avoiding multi-line escaping issues.
GCP_CREDENTIALS_FILE="${GCP_CREDENTIALS_FILE:-/path/to/your/gcp-service-account.json}"
GOOGLE_CLOUD_PROJECT="${GOOGLE_CLOUD_PROJECT:-REPLACE_WITH_GCP_PROJECT_ID}"
EMAIL_FROM="${EMAIL_FROM:-Sethum Methsanda <cv@sethum.dev>}"
ADMIN_EMAIL="${ADMIN_EMAIL:-sethummethsanda@gmail.com}"
GCP_STORAGE_BUCKET="${GCP_STORAGE_BUCKET:-portfolio-resume-assets}"
TURNSTILE_SITE_KEY="${TURNSTILE_SITE_KEY:-REPLACE_WITH_YOUR_TURNSTILE_SITE_KEY}"
RECAPTCHA_SITE_KEY="${RECAPTCHA_SITE_KEY:-REPLACE_WITH_YOUR_RECAPTCHA_SITE_KEY}"
RECAPTCHA_MIN_SCORE="${RECAPTCHA_MIN_SCORE:-0.5}"
NVIDIA_API_KEY_BACKUP="${NVIDIA_API_KEY_BACKUP:-REPLACE_WITH_YOUR_NVIDIA_BACKUP_KEY}"
SITE_URL="https://${DOMAIN}"


# =============================================================================
banner "STEP 1: Register Providers"
# =============================================================================
info "Registering Microsoft.App provider (Container Apps)…"
az provider register --namespace Microsoft.App
info "Registering Microsoft.OperationalInsights provider (Log Analytics)…"
az provider register --namespace Microsoft.OperationalInsights
info "Waiting for provider registration (polling every 15s, up to 10 min)…"
for i in $(seq 1 40); do
  APP_STATE=$(az provider show -n Microsoft.App --query registrationState -o tsv 2>/dev/null)
  OI_STATE=$(az provider show -n Microsoft.OperationalInsights --query registrationState -o tsv 2>/dev/null)
  if [[ "${APP_STATE}" == "Registered" && "${OI_STATE}" == "Registered" ]]; then
    break
  fi
  echo "  [${i}/40] Microsoft.App=${APP_STATE}  Microsoft.OperationalInsights=${OI_STATE} — waiting 15s…"
  sleep 15
done
APP_STATE=$(az provider show -n Microsoft.App --query registrationState -o tsv 2>/dev/null)
OI_STATE=$(az provider show -n Microsoft.OperationalInsights --query registrationState -o tsv 2>/dev/null)
if [[ "${APP_STATE}" != "Registered" || "${OI_STATE}" != "Registered" ]]; then
  warn "Providers not yet Registered (App=${APP_STATE}, OI=${OI_STATE}). You may need to wait and re-run from STEP 3."
else
  success "Providers registered."
fi


# =============================================================================
banner "STEP 2: Install containerapp CLI extension"
# =============================================================================
az extension add --name containerapp --upgrade -y 2>/dev/null || true
success "containerapp extension ready."


# =============================================================================
banner "STEP 3: Create Resource Group"
# =============================================================================
if az group show --name "${RESOURCE_GROUP}" &>/dev/null; then
  success "Resource group '${RESOURCE_GROUP}' already exists."
else
  info "Creating resource group '${RESOURCE_GROUP}' in '${LOCATION}'…"
  az group create \
    --name "${RESOURCE_GROUP}" \
    --location "${LOCATION}" \
    --output none
  success "Resource group created."
fi


# =============================================================================
banner "STEP 4: Create Container Apps Environment"
# =============================================================================
if az containerapp env show --name "${ENVIRONMENT}" --resource-group "${RESOURCE_GROUP}" &>/dev/null; then
  success "Container Apps environment '${ENVIRONMENT}' already exists."
else
  info "Creating Container Apps environment '${ENVIRONMENT}'…"
  az containerapp env create \
    --name "${ENVIRONMENT}" \
    --resource-group "${RESOURCE_GROUP}" \
    --location "${LOCATION}" \
    --logs-destination none \
    --output none
  success "Environment created."
fi

info "Fetching environment static IP (needed for apex A record)…"
STATIC_IP=""
for attempt in $(seq 1 10); do
  STATIC_IP="$(az containerapp env show \
    --name "${ENVIRONMENT}" \
    --resource-group "${RESOURCE_GROUP}" \
    --query "properties.staticIp" -o tsv 2>/dev/null || true)"
  if [[ -n "${STATIC_IP}" && "${STATIC_IP}" != "null" ]]; then
    break
  fi
  sleep 3
done
success "Static IP: ${STATIC_IP}"


# =============================================================================
banner "STEP 5: Create Container App (placeholder image)"
# =============================================================================
if az containerapp show --name "${APP_NAME}" --resource-group "${RESOURCE_GROUP}" &>/dev/null; then
  success "Container App '${APP_NAME}' already exists."
else
  info "Deploying Container App '${APP_NAME}' with placeholder image…"
  az containerapp create \
    --name "${APP_NAME}" \
    --resource-group "${RESOURCE_GROUP}" \
    --environment "${ENVIRONMENT}" \
    --image "${PLACEHOLDER_IMAGE}" \
    --target-port "${TARGET_PORT}" \
    --ingress external \
    --min-replicas 0 \
    --max-replicas 3 \
    --cpu 0.5 \
    --memory 1.0Gi \
    --output none
  success "Container App created with placeholder image."
fi


# =============================================================================
banner "STEP 6: Create Native Azure Secrets"
# Secrets are stored in ACA's built-in secret store — never in Activity Logs
# or plain-text ARM env vars.
# =============================================================================
info "Reading and single-lining GCP credentials JSON with jq…"
if [[ -f "${GCP_CREDENTIALS_FILE}" ]]; then
  GCP_CREDS_SINGLE_LINE="$(jq -c . "${GCP_CREDENTIALS_FILE}")"
  success "GCP credentials loaded and compacted."
else
  warn "GCP credentials file not found at '${GCP_CREDENTIALS_FILE}'."
  warn "Storing a placeholder. Update the Azure secret manually after setup."
  GCP_CREDS_SINGLE_LINE="REPLACE_WITH_BASE64_OR_SINGLE_LINE_GCP_JSON"
fi

info "Setting Azure Container App secrets…"
az containerapp secret set \
  --name "${APP_NAME}" \
  --resource-group "${RESOURCE_GROUP}" \
  --secrets \
    "supabase-db-url=${SUPABASE_DB_URL}" \
    "supabase-anon-key=${SUPABASE_ANON_KEY}" \
    "supabase-jwt-secret=${SUPABASE_JWT_SECRET}" \
    "supabase-service-role-key=${SUPABASE_SERVICE_ROLE_KEY}" \
    "supabase-url=${SUPABASE_URL}" \
    "resend-api-key=${RESEND_API_KEY}" \
    "turnstile-secret-key=${TURNSTILE_SECRET_KEY}" \
    "nvidia-nim-api=${NVIDIA_NIM_API}" \
    "nvidia-api-key-backup=${NVIDIA_API_KEY_BACKUP}" \
    "gcp-credentials=${GCP_CREDS_SINGLE_LINE}" \
  --output none
success "Secrets stored in Azure Container Apps native secret store."


# =============================================================================
banner "STEP 7: Map Environment Variables → Secret References"
# secretref: keeps env vars out of ARM definition entirely.
# Non-sensitive values are set directly.
# =============================================================================
info "Mapping env vars to secretrefs + setting non-sensitive vars…"
az containerapp update \
  --name "${APP_NAME}" \
  --resource-group "${RESOURCE_GROUP}" \
  --set-env-vars \
    "SUPABASE_DB_URL=secretref:supabase-db-url" \
    "SUPABASE_ANON_KEY=secretref:supabase-anon-key" \
    "SUPABASE_JWT_SECRET=secretref:supabase-jwt-secret" \
    "SUPABASE_SERVICE_ROLE_KEY=secretref:supabase-service-role-key" \
    "SUPABASE_URL=secretref:supabase-url" \
    "RESEND_API_KEY=secretref:resend-api-key" \
    "TURNSTILE_SECRET_KEY=secretref:turnstile-secret-key" \
    "NVIDIA_NIM_API=secretref:nvidia-nim-api" \
    "NVIDIA_API_KEY_BACKUP=secretref:nvidia-api-key-backup" \
    "GCP_CREDENTIALS=secretref:gcp-credentials" \
    "GOOGLE_CLOUD_PROJECT=${GOOGLE_CLOUD_PROJECT}" \
    "GOOGLE_CLOUD_LOCATION=us-central1" \
    "GOOGLE_GENAI_USE_VERTEXAI=True" \
    "ADMIN_EMAIL=${ADMIN_EMAIL}" \
    "EMAIL_FROM=${EMAIL_FROM}" \
    "EMAIL_PROVIDER=resend" \
    "GCP_STORAGE_BUCKET=${GCP_STORAGE_BUCKET}" \
    "TURNSTILE_SITE_KEY=${TURNSTILE_SITE_KEY}" \
    "RECAPTCHA_SITE_KEY=${RECAPTCHA_SITE_KEY}" \
    "RECAPTCHA_MIN_SCORE=${RECAPTCHA_MIN_SCORE}" \
    "SITE_URL=${SITE_URL}" \
    "ADDITIONAL_REDIRECT_URL_1=https://${DOMAIN}/login" \
    "PORT=${TARGET_PORT}" \
  --output none
success "Environment variables mapped (secrets via secretref, non-sensitive inline)."


# =============================================================================
# =============================================================================
banner "STEP 8: Configure GitHub Actions OIDC Authentication (Zero-Secret)"
# =============================================================================
IDENTITY_NAME="id-github-actions"
info "Creating User-Assigned Managed Identity '${IDENTITY_NAME}'…"
if az identity show --name "${IDENTITY_NAME}" --resource-group "${RESOURCE_GROUP}" &>/dev/null; then
  success "Managed Identity '${IDENTITY_NAME}' already exists."
else
  az identity create \
    --name "${IDENTITY_NAME}" \
    --resource-group "${RESOURCE_GROUP}" \
    --location "${LOCATION}" \
    --output none
  success "Managed Identity created."
fi

CLIENT_ID="$(az identity show --name "${IDENTITY_NAME}" --resource-group "${RESOURCE_GROUP}" --query clientId -o tsv)"
PRINCIPAL_ID="$(az identity show --name "${IDENTITY_NAME}" --resource-group "${RESOURCE_GROUP}" --query principalId -o tsv)"
TENANT_ID="$(az account show --query tenantId -o tsv)"

info "Assigning Contributor role to Managed Identity on resource group…"
az role assignment create \
  --assignee-object-id "${PRINCIPAL_ID}" \
  --assignee-principal-type ServicePrincipal \
  --role Contributor \
  --scope "/subscriptions/${SUBSCRIPTION_ID}/resourceGroups/${RESOURCE_GROUP}" \
  --output none 2>/dev/null || true
success "Role assigned."

info "Creating Federated Identity Credential for GitHub Actions (OIDC)…"
az identity federated-credential create \
  --name fc-github-main \
  --identity-name "${IDENTITY_NAME}" \
  --resource-group "${RESOURCE_GROUP}" \
  --issuer "https://token.actions.githubusercontent.com" \
  --subject "repo:sethum-VS/my-portfolio:ref:refs/heads/main" \
  --audiences "api://AzureADTokenExchange" \
  --output none 2>/dev/null || true
success "OIDC Federated Credential configured."

# Get the App FQDN
APP_FQDN="$(az containerapp show \
  --name "${APP_NAME}" \
  --resource-group "${RESOURCE_GROUP}" \
  --query "properties.configuration.ingress.fqdn" -o tsv)"

# Get domain verification ID
DOMAIN_VERIFICATION_ID="$(az containerapp show \
  --name "${APP_NAME}" \
  --resource-group "${RESOURCE_GROUP}" \
  --query "properties.customDomainVerificationId" -o tsv)"


# =============================================================================
banner "SETUP COMPLETE — GitHub Secrets & DNS Configured"
# =============================================================================

echo -e "\n${BOLD}${GREEN}════════════════════════════════════════════${RESET}"
echo -e "${BOLD}${GREEN}  ✅  Infrastructure provisioned successfully!${RESET}"
echo -e "${BOLD}${GREEN}════════════════════════════════════════════${RESET}\n"

echo -e "${BOLD}Container App FQDN (placeholder image):${RESET}"
echo -e "  https://${APP_FQDN}\n"

# ─── GitHub Secrets ───────────────────────────────────────────────────────────
echo -e "${BOLD}${YELLOW}────────────────────────────────────────────${RESET}"
echo -e "${BOLD}${YELLOW}  GitHub Actions OIDC Secrets (Already set via gh CLI):${RESET}"
echo -e "${BOLD}${YELLOW}────────────────────────────────────────────${RESET}"
echo -e "  AZURE_CLIENT_ID       : ${BOLD}${CLIENT_ID}${RESET}"
echo -e "  AZURE_TENANT_ID       : ${BOLD}${TENANT_ID}${RESET}"
echo -e "  AZURE_SUBSCRIPTION_ID : ${BOLD}${SUBSCRIPTION_ID}${RESET}\n"

# ─── DNS ──────────────────────────────────────────────────────────────────────
echo -e "${BOLD}${YELLOW}────────────────────────────────────────────${RESET}"
echo -e "${BOLD}${YELLOW}  ACTION 2: Add DNS Records for sethum.dev${RESET}"
echo -e "${BOLD}${YELLOW}────────────────────────────────────────────${RESET}"
echo -e "Add these records at your DNS provider BEFORE running the domain binding:"
echo ""
echo -e "  ${BOLD}Record 1 — Ownership Verification (TXT)${RESET}"
echo -e "  Type : TXT"
echo -e "  Name : asuid.${DOMAIN}  (or asuid.@ )"
echo -e "  Value: ${BOLD}${DOMAIN_VERIFICATION_ID}${RESET}"
echo ""
echo -e "  ${BOLD}Record 2 — Apex Domain Routing (A)${RESET}"
echo -e "  ${RED}⚠ CNAME is NOT supported on apex domains. Use an A record (or ALIAS/ANAME if your provider supports it).${RESET}"
echo -e "  Type : A"
echo -e "  Name : ${DOMAIN}  (or @ )"
echo -e "  Value: ${BOLD}${STATIC_IP}${RESET}"
echo ""
echo -e "Wait for DNS propagation (typically 2–15 minutes), then run ACTION 3.\n"

# ─── Domain binding ───────────────────────────────────────────────────────────
echo -e "${BOLD}${YELLOW}────────────────────────────────────────────${RESET}"
echo -e "${BOLD}${YELLOW}  ACTION 3: Bind Domain + Provision TLS Cert${RESET}"
echo -e "${BOLD}${YELLOW}────────────────────────────────────────────${RESET}"
echo -e "After DNS propagates, run these two commands:\n"

echo -e "${BOLD}# Step A — Bind the custom domain (validates DNS, sets up routing):${RESET}"
cat <<EOF
az containerapp hostname bind \\
  --name ${APP_NAME} \\
  --resource-group ${RESOURCE_GROUP} \\
  --hostname ${DOMAIN} \\
  --environment ${ENVIRONMENT} \\
  --validation-method HTTP
EOF

echo ""
echo -e "${BOLD}# Step B — Issue the free managed TLS certificate:${RESET}"
cat <<EOF
az containerapp ssl upload \\
  --name ${APP_NAME} \\
  --resource-group ${RESOURCE_GROUP} \\
  --hostname ${DOMAIN} \\
  --environment ${ENVIRONMENT} \\
  --certificate-file ""
EOF
echo -e "${CYAN}(Leaving --certificate-file blank triggers the free managed certificate.)${RESET}\n"

echo -e "${BOLD}${GREEN}All done! Push to main to trigger the GitHub Actions deployment.${RESET}\n"
