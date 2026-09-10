package awsysco

import "net/url"

// Static (non-parameterized) endpoint paths.
const (
	pathLinks                 = "/api/v1/links"
	pathFoldersV1             = "/api/v1/folders"
	pathMe                    = "/api/v1/me"
	pathBulk                  = "/api/v1/bulk"
	pathImports               = "/api/v1/imports"
	pathUserDomains           = "/api/user/domains"
	pathUserNamespace         = "/api/user/namespace"
	pathUserStats             = "/api/user/stats"
	pathUserProfile           = "/api/user/profile"
	pathUserUtmTemplates      = "/api/user/utm-templates"
	pathRecentClicks          = "/api/user/clicks/recent"
	pathWebhooksV1            = "/api/v1/webhooks"
	pathWebhookEventTypes     = "/api/webhooks/event-types"
	pathViews                 = "/api/views"
	pathAffiliatePrograms     = "/api/affiliate/programs"
	pathAffiliateDiscover     = "/api/affiliate/discover"
	pathAffiliatePartnerships = "/api/affiliate/partnerships"
	pathAffiliateLimits       = "/api/affiliate/limits"
	pathAgentlinkSubscribe    = "/api/agentlink/subscribe"
	pathAgentlinkAccountStats = "/api/agentlink/account/stats"
	pathExportLinks           = "/api/export/links"
)

// Links.
func pathLink(id string) string                      { return pathLinks + "/" + url.PathEscape(id) }
func pathLinkStats(shortPath string) string          { return pathLink(shortPath) + "/stats" }
func pathLinkAggregateStats(shortPath string) string { return pathLink(shortPath) + "/stats/aggregate" }
func pathLinkFolder(linkID string) string            { return pathLink(linkID) + "/folder" }

// Tags / QR (mounted under the singular /api/link/{shortPath} prefix, distinct from /api/v1/links).
func pathTags(shortPath string) string     { return "/api/link/" + url.PathEscape(shortPath) + "/tags" }
func pathTag(shortPath, tag string) string { return pathTags(shortPath) + "/" + url.PathEscape(tag) }
func pathQRSettings(shortPath string) string {
	return "/api/link/" + url.PathEscape(shortPath) + "/qr-settings"
}
func pathQRImage(shortCode string) string  { return "/api/qr/" + url.PathEscape(shortCode) }
func pathLinkScan(shortPath string) string { return "/api/link-scan/" + url.PathEscape(shortPath) }
func pathExportStats(shortPath string) string {
	return "/api/export/stats/" + url.PathEscape(shortPath)
}

// Folders. NOTE: the platform is inconsistent here — List/Create/Delete are
// mounted under /api/v1/folders, but Update only exists at the unversioned
// /api/folders/{id} twin (the /api/v1 PATCH route 404s live; see ADR-011).
// pathFolderV1 and pathFolder are deliberately different paths, not a typo.
func pathFolderV1(id string) string { return pathFoldersV1 + "/" + url.PathEscape(id) }
func pathFolder(id string) string   { return "/api/folders/" + url.PathEscape(id) }

// Domains / namespace.
func pathUserDomain(domain string) string       { return pathUserDomains + "/" + url.PathEscape(domain) }
func pathUserDomainVerify(domain string) string { return pathUserDomain(domain) + "/verify" }
func pathDomainCheck(hostname string) string    { return "/api/domains/check/" + url.PathEscape(hostname) }
func pathNamespaceCheck(ns string) string       { return "/api/namespace/check/" + url.PathEscape(ns) }

// Webhooks / saved views / UTM templates. NOTE: like folders, the platform is
// inconsistent here — List/Create/Delete/Test are mounted under
// /api/v1/webhooks, but Update only exists at the unversioned
// /api/webhooks/{id} twin. pathWebhookV1 and pathWebhookUpdate are
// deliberately different paths, not a typo.
func pathWebhookV1(id string) string     { return pathWebhooksV1 + "/" + url.PathEscape(id) }
func pathWebhookTest(id string) string   { return pathWebhookV1(id) + "/test" }
func pathWebhookUpdate(id string) string { return "/api/webhooks/" + url.PathEscape(id) }
func pathView(id string) string          { return pathViews + "/" + url.PathEscape(id) }
func pathUtmTemplate(id string) string   { return pathUserUtmTemplates + "/" + url.PathEscape(id) }

// Affiliate.
func pathAffiliateProgram(id string) string         { return pathAffiliatePrograms + "/" + url.PathEscape(id) }
func pathAffiliateProgramStats(id string) string    { return pathAffiliateProgram(id) + "/stats" }
func pathAffiliateProgramPartners(id string) string { return pathAffiliateProgram(id) + "/partners" }
func pathAffiliateProgramPartner(programID, partnerID string) string {
	return pathAffiliateProgramPartners(programID) + "/" + url.PathEscape(partnerID)
}
func pathAffiliateJoin(programID string) string {
	return "/api/affiliate/join/" + url.PathEscape(programID)
}
func pathAffiliatePartnership(id string) string {
	return pathAffiliatePartnerships + "/" + url.PathEscape(id)
}
func pathAffiliatePartnershipStats(id string) string { return pathAffiliatePartnership(id) + "/stats" }

// Agentlink.
func pathAgentlinkLinkStats(shortPath string) string {
	return "/api/agentlink/links/" + url.PathEscape(shortPath) + "/stats"
}

// Imports / Web2App.
func pathImport(jobID string) string                { return pathImports + "/" + url.PathEscape(jobID) }
func pathImportRedirectMapCSV(jobID string) string  { return pathImport(jobID) + "/redirect-map.csv" }
func pathImportRedirectMapJSON(jobID string) string { return pathImport(jobID) + "/redirect-map.json" }
func pathWeb2App(token string) string               { return "/api/v1/web2app/" + url.PathEscape(token) }
